package kcache

import (
	"context"
	builtin_errors "errors"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/boz/kcache/filter"
	"github.com/pkg/errors"
)

var (
	ErrNotRunning = builtin_errors.New("Not running")
)

type Publisher interface {
	Subscribe() (Subscription, error)
	SubscribeWithFilter(filter.Filter) (FilterSubscription, error)
	SubscribeForFilter() (FilterSubscription, error)
	Clone() (Controller, error)
	CloneWithFilter(filter.Filter) (FilterController, error)
	CloneForFilter() (FilterController, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Controller interface {
	CacheController
	Publisher
	Done() <-chan struct{}
	Close()
	Error() error
}

func NewController(ctx context.Context, log logutil.Log, client client.Client) (Controller, error) {
	return NewBuilder().
		Context(ctx).
		Log(log).
		Client(client).
		Create()
}

type controller struct {

	// closed when initialization complete
	readych chan struct{}

	watcher watcher
	lister  lister
	cache   cache

	subscription subscription
	publisher    Publisher

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func (c *controller) Ready() <-chan struct{} {
	return c.readych
}

func (c *controller) Close() {
	c.lc.Shutdown(nil)
}

func (c *controller) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *controller) Error() error {
	return c.lc.Error()
}

func (c *controller) Cache() CacheReader {
	return c.cache
}

func (c *controller) Subscribe() (Subscription, error) {
	return c.publisher.Subscribe()
}

func (c *controller) SubscribeWithFilter(f filter.Filter) (FilterSubscription, error) {
	return c.publisher.SubscribeWithFilter(f)
}

func (c *controller) SubscribeForFilter() (FilterSubscription, error) {
	return c.publisher.SubscribeForFilter()
}

func (c *controller) Clone() (Controller, error) {
	return c.publisher.Clone()
}

func (c *controller) CloneWithFilter(f filter.Filter) (FilterController, error) {
	return c.publisher.CloneWithFilter(f)
}

func (c *controller) CloneForFilter() (FilterController, error) {
	return c.publisher.CloneForFilter()
}

func (c *controller) run() {
	defer c.lc.ShutdownCompleted()
	initialized := false

mainloop:
	for {
		select {

		case err := <-c.lc.ShutdownRequest():

			c.log.Debugf("shutdown request: %v", err)
			c.lc.ShutdownInitiated(err)
			break mainloop

		case <-c.lister.Done():

			err := c.lister.Error()
			c.log.Debugf("lister complete: %v", err)
			c.lc.ShutdownInitiated(errors.Wrap(err, "lister complete"))
			break mainloop

		case <-c.watcher.Done():

			err := c.watcher.Error()
			c.log.Debugf("watcher complete: %v", err)
			c.lc.ShutdownInitiated(errors.Wrap(err, "watcher complete"))
			break mainloop

		case <-c.cache.Done():

			err := c.cache.Error()
			c.log.Debugf("cache complete: %v", err)
			c.lc.ShutdownInitiated(errors.Wrap(err, "cache complete"))
			break mainloop

		case result := <-c.lister.Result():

			if result.err != nil {
				c.log.Errorf("lister error: %v", result.err)
				c.lc.ShutdownInitiated(errors.Wrap(result.err, "lister result"))
				break mainloop
			}

			version, err := listResourceVersion(result.list)
			if err != nil {
				c.log.Errorf("resource version error: %v", err)
				c.lc.ShutdownInitiated(errors.Wrap(err, "listing resource version"))
				break mainloop
			}

			c.log.Debugf("list version: %v", version)

			list, err := extractList(result.list)
			if err != nil {
				c.log.Errorf("extract list error: %v", err)
				c.lc.ShutdownInitiated(errors.Wrap(err, "extracting list"))
				break mainloop
			}

			events, err := c.cache.sync(list)
			if err != nil {
				c.log.Errorf("cache sync error: %v", err)
				c.lc.ShutdownInitiated(err)
				break mainloop
			}

			c.log.Debugf("list complete: version: %v, items: %v, events: %v",
				version, len(list), len(events))

			if !initialized {
				c.log.Debugf("ready")
				initialized = true
				close(c.readych)
			} else {
				c.distributeEvents(events)
			}

			if err := c.watcher.reset(version); err != nil {
				c.log.Errorf("watcher reset error: %v", err)
				c.lc.ShutdownInitiated(errors.Wrap(err, "watcher reset"))
				break mainloop
			}

		case evt := <-c.watcher.events():
			c.log.Debugf("update event: %v", evt)

			events, err := c.cache.update(evt)
			if err != nil {
				c.log.Errorf("update event: cache update error %v", err)
				c.lc.ShutdownInitiated(errors.Wrap(err, "updating cache"))
				break mainloop
			}
			c.distributeEvents(events)
		}
	}

	<-c.cache.Done()
	<-c.watcher.Done()
	<-c.lister.Done()
}

func (c *controller) distributeEvents(events []Event) {
	for _, evt := range events {
		c.subscription.send(evt)
	}
	c.log.Debugf("distribute events: %v events", len(events))
}
