package kcache

import (
	"context"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/pkg/errors"
)

type FilterSubscription interface {
	Subscription
	Refilter(filter.Filter) error
}

type filterSubscription struct {
	parent Subscription

	deferReady bool
	refilterch chan filter.Filter

	outch   chan Event
	readych chan struct{}

	filter filter.Filter
	cache  cache

	lc  lifecycle.Lifecycle
	log logutil.Log
}

func newFilterSubscription(log logutil.Log, parent Subscription, f filter.Filter, deferReady bool) FilterSubscription {

	ctx := context.Background()
	lc := lifecycle.New()

	s := &filterSubscription{
		parent:     parent,
		refilterch: make(chan filter.Filter),
		outch:      make(chan Event, EventBufsiz),
		readych:    make(chan struct{}),
		deferReady: deferReady,
		filter:     f,
		cache:      newCache(ctx, log, lc.ShuttingDown(), f),
		lc:         lc,
		log:        log,
	}

	go s.run()

	return s
}

func (s *filterSubscription) Cache() CacheReader {
	return s.cache
}
func (s *filterSubscription) Ready() <-chan struct{} {
	return s.readych
}
func (s *filterSubscription) Events() <-chan Event {
	return s.outch
}
func (s *filterSubscription) Close() {
	s.parent.Close()
}
func (s *filterSubscription) Done() <-chan struct{} {
	return s.lc.Done()
}
func (s *filterSubscription) Error() error {
	if err := s.lc.Error(); err != nil {
		return err
	}
	return s.parent.Error()
}

func (s *filterSubscription) Refilter(filter filter.Filter) error {
	select {
	case s.refilterch <- filter:
		return nil
	case <-s.lc.ShuttingDown():
		return errors.WithStack(ErrNotRunning)
	}
}

func (s *filterSubscription) run() {
	defer s.lc.ShutdownCompleted()

	preadych := s.parent.Ready()

	pending := false
	ready := false

loop:
	for {
		select {
		case err := <-s.lc.ShutdownRequest():
			s.log.Debugf("shutdown requested: %v", err)
			s.lc.ShutdownInitiated(err)
			break loop

		case <-preadych:

			preadych = nil

			if s.deferReady && !pending {
				s.log.Debugf("parent ready: deferring ready")
				continue
			}

			list, err := s.parent.Cache().List()
			if err != nil {
				s.log.Debugf("parent ready: cache list error: %v", err)
				s.lc.ShutdownInitiated(errors.Wrap(err, "parent ready: cache list"))
				break loop
			}

			if _, err := s.cache.sync(list); err != nil {
				s.log.Debugf("parent ready: cache sync error: %v", err)
				s.lc.ShutdownInitiated(errors.Wrap(err, "parent ready: cache sync"))
				break loop
			}

			s.log.Debugf("parent ready: making ready")
			close(s.readych)
			ready = true

		case f := <-s.refilterch:
			s.log.Debugf("refiltering...")

			isNew := !filter.FiltersEqual(s.filter, f)

			switch {

			case preadych != nil && !isNew:
				s.log.Debugf("refilter: deferring ready (filter unchanged)")
				pending = true
				continue

			case preadych != nil && isNew:
				if _, err := s.cache.refilter(nil, f); err != nil {
					s.log.Debugf("refilter: cache refilter (not ready): %v", err)
					s.lc.ShutdownInitiated(errors.Wrap(err, "refilter: cache refilter (not ready)"))
					break loop
				}
				s.log.Debugf("refilter: deferring ready (filter changed)")
				s.filter = f
				pending = true
				continue

			case ready && !isNew:
				s.log.Debugf("refilter: filter unchanged")
				continue

			case !ready && !isNew:
				s.log.Debugf("refilter: making ready (filter unchanged)")
				close(s.readych)
				ready = true
				continue

			}

			// pready == nil && isNew

			list, err := s.parent.Cache().List()
			if err != nil {
				s.log.Debugf("refilter: cache list error: %v", err)
				s.lc.ShutdownInitiated(errors.Wrap(err, "refilter: cache list"))
				break loop
			}

			events, err := s.cache.refilter(list, f)
			if err != nil {
				s.log.Debugf("refilter: cache refilter error: %v", err)
				s.lc.ShutdownInitiated(errors.Wrap(err, "refilter: cache refilter"))
				break loop
			}
			s.filter = f

			if !ready {
				s.log.Debugf("refilter: making ready (filter changed)")
				close(s.readych)
				ready = true
				continue
			}

			s.log.Debugf("refilter: %v events", len(events))

			s.distributeEvents(events)

		case evt, ok := <-s.parent.Events():

			switch {
			case !ok:
				s.log.Debugf("update: parent closed")
				s.lc.ShutdownInitiated(nil)
				break loop
			case !ready:
				continue
			}

			events, err := s.cache.update(evt)
			if err != nil {
				s.log.Debugf("update: cache update error %v", err)
				s.lc.ShutdownInitiated(nil)
				break loop
			}

			s.log.Debugf("update: %v events", len(events))

			s.distributeEvents(events)

		}
	}

	s.parent.Close()

	close(s.outch)

	<-s.parent.Done()
}

func (s *filterSubscription) distributeEvents(events []Event) {
	for _, evt := range events {
		select {
		case s.outch <- evt:
		default:
			s.log.Warnf("event buffer overrun")
		}
	}
}
