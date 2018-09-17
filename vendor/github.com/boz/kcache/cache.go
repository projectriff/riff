package kcache

import (
	"context"
	"strconv"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CacheReader interface {
	GetObject(obj metav1.Object) (metav1.Object, error)
	Get(ns string, name string) (metav1.Object, error)
	List() ([]metav1.Object, error)
}

type cache interface {
	CacheReader
	sync([]metav1.Object) ([]Event, error)
	update(Event) ([]Event, error)
	refilter([]metav1.Object, filter.Filter) ([]Event, error)
	Done() <-chan struct{}
	Error() error
}

type cacheKey struct {
	namespace string
	name      string
}

type cacheEntry struct {
	version int
	object  metav1.Object
}

type syncRequest struct {
	list     []metav1.Object
	resultch chan<- []Event
}

type getRequest struct {
	key      cacheKey
	resultch chan<- metav1.Object
}

type updateRequest struct {
	evt      Event
	resultch chan<- []Event
}

type refilterRequest struct {
	list     []metav1.Object
	filter   filter.Filter
	resultch chan<- []Event
}

type _cache struct {
	filter     filter.Filter
	syncch     chan syncRequest
	updatech   chan updateRequest
	refilterch chan refilterRequest

	getch  chan getRequest
	listch chan chan []metav1.Object

	items map[cacheKey]cacheEntry

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func newCache(ctx context.Context, log logutil.Log, stopch <-chan struct{}, filter filter.Filter) cache {
	log = log.WithComponent("cache")

	c := &_cache{
		filter:     filter,
		syncch:     make(chan syncRequest),
		updatech:   make(chan updateRequest),
		getch:      make(chan getRequest),
		refilterch: make(chan refilterRequest),
		listch:     make(chan chan []metav1.Object),
		items:      make(map[cacheKey]cacheEntry),
		log:        log,
		lc:         lifecycle.New(),
		ctx:        ctx,
	}

	go c.lc.WatchContext(ctx)
	go c.lc.WatchChannel(stopch)
	go c.run()

	return c
}

func (c *_cache) sync(list []metav1.Object) ([]Event, error) {
	resultch := make(chan []Event, 1)
	request := syncRequest{list, resultch}

	select {
	case <-c.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case c.syncch <- request:
	}

	return <-resultch, nil
}

func (c *_cache) update(evt Event) ([]Event, error) {

	resultch := make(chan []Event, 1)
	request := updateRequest{evt, resultch}

	select {
	case <-c.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case c.updatech <- request:
	}

	return <-resultch, nil

}

func (c *_cache) refilter(list []metav1.Object, filter filter.Filter) ([]Event, error) {
	resultch := make(chan []Event, 1)
	request := refilterRequest{list, filter, resultch}

	select {
	case <-c.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case c.refilterch <- request:
	}

	return <-resultch, nil
}

func (c *_cache) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *_cache) Error() error {
	return c.lc.Error()
}

func (c *_cache) List() ([]metav1.Object, error) {
	resultch := make(chan []metav1.Object, 1)

	select {
	case <-c.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case c.listch <- resultch:
	}

	return <-resultch, nil
}

func (c *_cache) GetObject(obj metav1.Object) (metav1.Object, error) {
	return c.Get(obj.GetNamespace(), obj.GetName())
}

func (c *_cache) Get(ns, name string) (metav1.Object, error) {
	resultch := make(chan metav1.Object, 1)
	key := cacheKey{ns, name}
	request := getRequest{key, resultch}
	select {
	case <-c.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case c.getch <- request:
	}
	return <-resultch, nil
}

func (c *_cache) run() {
	defer c.lc.ShutdownCompleted()
	for {
		select {
		case request := <-c.syncch:
			request.resultch <- c.doSync(request.list)
		case request := <-c.updatech:
			request.resultch <- c.doUpdate(request.evt)
		case request := <-c.refilterch:
			request.resultch <- c.doRefilter(request.list, request.filter)
		case request := <-c.listch:
			request <- c.doList()
		case request := <-c.getch:
			if entry, ok := c.items[request.key]; ok {
				request.resultch <- entry.object
			} else {
				request.resultch <- nil
			}
		case err := <-c.lc.ShutdownRequest():
			c.lc.ShutdownInitiated(err)
			return
		}
	}
}

func (c *_cache) doList() []metav1.Object {
	result := make([]metav1.Object, 0, len(c.items))
	for _, obj := range c.items {
		result = append(result, obj.object)
	}
	return result
}

func (c *_cache) doSync(list []metav1.Object) []Event {

	var events []Event
	set := make(map[cacheKey]cacheEntry)

	for _, obj := range list {

		key, err := c.createKey(obj)
		if err != nil {
			c.log.ErrWarn(err, "createKey(%T)", obj)
			continue
		}

		entry, err := c.createEntry(obj)
		if err != nil {
			c.log.ErrWarn(err, "createEntry(%T)", obj)
			continue
		}

		current, found := c.items[key]

		accept := c.filter.Accept(entry.object)

		switch {
		case accept && !found:
			events = append(events, NewEvent(EventTypeCreate, entry.object))
			c.items[key] = entry
		case accept && current.version < entry.version:
			events = append(events, NewEvent(EventTypeUpdate, entry.object))
			c.items[key] = entry
		case current.version >= entry.version:
			if !c.filter.Accept(current.object) {
				continue
			}
		default:
			// don't add to working new working set of objects
			continue
		}

		set[key] = entry
	}

	for k, current := range c.items {
		if _, ok := set[k]; !ok {
			events = append(events, NewEvent(EventTypeDelete, current.object))
			delete(c.items, k)
		}
	}

	return events
}

func (c *_cache) doRefilter(list []metav1.Object, filter filter.Filter) []Event {
	c.filter = filter
	return c.doSync(list)
}

func (c *_cache) doUpdate(evt Event) []Event {
	events := make([]Event, 0, 1)

	obj := evt.Resource()

	version, err := strconv.Atoi(obj.GetResourceVersion())
	if err != nil {
		c.log.ErrWarn(err, "resource version %v", obj.GetResourceVersion())
		return events
	}

	key := cacheKey{obj.GetNamespace(), obj.GetName()}
	entry := cacheEntry{version, obj}

	current, found := c.items[key]

	accept := c.filter.Accept(entry.object)

	switch evt.Type() {
	case EventTypeDelete:
		if found {
			events = append(events, evt)
			delete(c.items, key)
		}
	default:
		switch {
		case !accept && !found:
			// do nothing
		case accept && !found:
			// create
			events = append(events, NewEvent(EventTypeCreate, obj))
			c.items[key] = entry
		case accept && current.version < entry.version:
			// update
			events = append(events, NewEvent(EventTypeUpdate, obj))
			c.items[key] = entry
		case !accept && current.version < entry.version:
			// filter-delete
			events = append(events, NewEvent(EventTypeDelete, obj))
			delete(c.items, key)
		}
	}

	return events
}

func (c *_cache) createKey(obj metav1.Object) (cacheKey, error) {
	ns := obj.GetNamespace()
	name := obj.GetName()

	return cacheKey{ns, name}, nil
}

func (c *_cache) createEntry(obj metav1.Object) (cacheEntry, error) {
	version, err := strconv.Atoi(obj.GetResourceVersion())
	if err != nil {
		return cacheEntry{}, err
	}
	return cacheEntry{version, obj}, nil
}
