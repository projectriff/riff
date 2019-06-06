package kcache

import (
	"context"
	"time"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/pkg/errors"
)

const (
	watchRetryDelay = time.Second
)

type watcher interface {
	reset(string) error
	events() <-chan Event

	Done() <-chan struct{}
	Error() error
}

type _watcher struct {
	version string

	client client.WatchClient

	resetch chan string
	evtch   chan chan (<-chan Event)

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func newWatcher(ctx context.Context, log logutil.Log, stopch <-chan struct{}, client client.WatchClient) watcher {
	log = log.WithComponent("watcher")
	lc := lifecycle.New()

	w := &_watcher{
		client:  client,
		resetch: make(chan string),
		evtch:   make(chan chan (<-chan Event)),
		log:     log,
		lc:      lc,
		ctx:     ctx,
	}

	go w.lc.WatchContext(ctx)
	go w.lc.WatchChannel(stopch)
	go w.run()

	return w
}

func (w *_watcher) reset(vsn string) error {
	select {
	case w.resetch <- vsn:
		return nil
	case <-w.lc.ShuttingDown():
		return errors.WithStack(ErrNotRunning)
	}
}

func (w *_watcher) events() <-chan Event {
	req := make(chan (<-chan Event), 1)
	select {
	case w.evtch <- req:
		return <-req
	case <-w.lc.ShuttingDown():
		return nil
	}
}

func (w *_watcher) Done() <-chan struct{} {
	return w.lc.Done()
}

func (w *_watcher) Error() error {
	return w.lc.Error()
}

func (w *_watcher) run() {
	defer w.lc.ShutdownCompleted()

	ctx, cancel := context.WithCancel(w.ctx)
	defer cancel()

	var session watchSession = nullWatchSession{}
	var outch chan Event

	var curVersion string

	var retry *time.Timer

mainloop:
	for {

		select {
		case err := <-w.lc.ShutdownRequest():
			w.log.Debugf("shutdown request: %v", err)
			w.lc.ShutdownInitiated(err)
			break mainloop

		case vsn := <-w.resetch:
			w.log.Debugf("ressetting to version %v", vsn)

			if retry != nil {
				retry.Stop()
				retry = nil
			}

			session.stop()
			session = newWatchSession(ctx, w.log, w.client, vsn)
			outch = make(chan Event, EventBufsiz)
			curVersion = vsn

		case <-session.done():
			w.log.Debugf("session done.  retrying version %v in %v", curVersion, watchRetryDelay)

			session.stop()
			session = nullWatchSession{}
			outch = nil
			retry = w.scheduleRetry(w.resetch, curVersion)

		case evt := <-session.events():

			select {
			case outch <- evt:
			default:
				w.log.Errorf("output buffer full")
			}

			curVersion = evt.Resource().GetResourceVersion()

			w.log.Debugf("session event: %v version: %v", evt, curVersion)

		case reqch := <-w.evtch:
			reqch <- outch
		}
	}

	cancel()
	if retry != nil {
		retry.Stop()
	}

	if donech := session.done(); donech != nil {
		<-donech
	}
}

func (w *_watcher) scheduleRetry(ch chan string, vsn string) *time.Timer {
	return time.AfterFunc(watchRetryDelay, func() {
		select {
		case ch <- vsn:
		case <-w.lc.ShuttingDown():
		}
	})
}
