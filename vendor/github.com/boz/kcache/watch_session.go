package kcache

import (
	"context"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type watchSession interface {
	events() <-chan Event
	done() <-chan struct{}
	stop()
	Error() error
}

type nullWatchSession struct{}

func (nullWatchSession) events() <-chan Event  { return nil }
func (nullWatchSession) done() <-chan struct{} { return nil }
func (nullWatchSession) stop()                 {}
func (nullWatchSession) Error() error          { return nil }

type _watchSession struct {
	client  client.WatchClient
	version string

	outch chan Event

	ctx    context.Context
	cancel context.CancelFunc
	log    logutil.Log
	lc     lifecycle.Lifecycle
}

func newWatchSession(ctx context.Context, log logutil.Log, client client.WatchClient, version string) watchSession {
	lc := lifecycle.New()

	ctx, cancel := context.WithCancel(ctx)

	s := &_watchSession{
		client:  client,
		version: version,
		outch:   make(chan Event, EventBufsiz),
		ctx:     ctx,
		cancel:  cancel,
		log:     log.WithComponent("watch-session"),
		lc:      lc,
	}

	go lc.WatchContext(ctx)
	go s.run()
	return s
}

func (s *_watchSession) done() <-chan struct{} {
	return s.lc.Done()
}

func (s *_watchSession) stop() {
	s.lc.ShutdownAsync(nil)
}

func (s *_watchSession) Error() error {
	return s.lc.Error()
}

func (s *_watchSession) events() <-chan Event {
	return s.outch
}

func (s *_watchSession) run() {
	defer s.lc.ShutdownCompleted()
	defer s.cancel()

	conn, err := s.connect()
	if err != nil {
		s.log.Debugf("connecting to server: %v", err)
		s.lc.ShutdownInitiated(errors.Wrap(err, "connecting to server"))
		return
	}

	defer conn.Stop()

	for {
		select {

		case err := <-s.lc.ShutdownRequest():

			s.lc.ShutdownInitiated(err)
			return

		case kevt, ok := <-conn.ResultChan():

			if !ok {
				s.lc.ShutdownInitiated(nil)
				return
			}

			if status, ok := kevt.Object.(*metav1.Status); ok {
				s.logStatus(status)
				continue
			}

			obj, err := meta.Accessor(kevt.Object)
			if err != nil {
				s.lc.ShutdownInitiated(errors.Wrap(err, "meta accessor"))
				return
			}

			var evt Event

			switch kevt.Type {
			case watch.Added:
				evt = NewEvent(EventTypeCreate, obj)
			case watch.Modified:
				evt = NewEvent(EventTypeUpdate, obj)
			case watch.Deleted:
				evt = NewEvent(EventTypeDelete, obj)
			}

			if evt == nil {
				s.log.Debugf("unknown event type: %v", kevt.Type)
				continue
			}

			select {
			case s.outch <- evt:
			default:
				s.log.Warnf("output buffer full; event missed.")
			}

		}
	}
}

func (s *_watchSession) connect() (watch.Interface, error) {
	response, err := s.client.Watch(s.ctx, metav1.ListOptions{
		ResourceVersion: s.version,
		Watch:           true,
	})
	return response, err
}

func (s *_watchSession) logStatus(status *metav1.Status) {
	s.log.Debugf("STATUS: %v %v %v [code: %v vsn: %v]", status.Status, status.Message, status.Reason, status.Code, status.GetResourceVersion())
}
