package kcache

import (
	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/pkg/errors"
)

type FilterController interface {
	Controller
	Refilter(filter.Filter) error
}

type publisher struct {
	parent Subscription

	subscribech   chan chan<- Subscription
	unsubscribech chan subscription
	subscriptions map[subscription]struct{}

	lc  lifecycle.Lifecycle
	log logutil.Log
}

func newPublisher(log logutil.Log, parent Subscription) Controller {
	s := &publisher{
		parent:        parent,
		subscribech:   make(chan chan<- Subscription),
		unsubscribech: make(chan subscription),
		subscriptions: make(map[subscription]struct{}),
		lc:            lifecycle.New(),
		log:           log.WithComponent("publisher"),
	}

	go s.run()

	return s
}

func (s *publisher) Ready() <-chan struct{} {
	return s.parent.Ready()
}

func (s *publisher) Cache() CacheReader {
	return s.parent.Cache()
}

func (s *publisher) Close() {
	s.parent.Close()
}

func (s *publisher) Done() <-chan struct{} {
	return s.lc.Done()
}

func (s *publisher) Error() error {
	return s.lc.Error()
}

func (s *publisher) Subscribe() (Subscription, error) {
	resultch := make(chan Subscription, 1)
	select {
	case <-s.lc.ShuttingDown():
		return nil, errors.WithStack(ErrNotRunning)
	case s.subscribech <- resultch:
		return <-resultch, nil
	}
}

func (s *publisher) SubscribeWithFilter(f filter.Filter) (FilterSubscription, error) {
	sub, err := s.Subscribe()
	if err != nil {
		return nil, err
	}
	return newFilterSubscription(s.log, sub, f, false), nil
}

func (s *publisher) SubscribeForFilter() (FilterSubscription, error) {
	sub, err := s.Subscribe()
	if err != nil {
		return nil, err
	}
	return newFilterSubscription(s.log, sub, filter.All(), true), nil
}

func (s *publisher) Clone() (Controller, error) {
	sub, err := s.Subscribe()
	if err != nil {
		return nil, err
	}
	return newPublisher(s.log, sub), nil
}

func (s *publisher) CloneWithFilter(f filter.Filter) (FilterController, error) {
	sub, err := s.SubscribeWithFilter(f)
	if err != nil {
		return nil, err
	}
	return newFilterPublisher(s.log, sub), nil
}

func (s *publisher) CloneForFilter() (FilterController, error) {
	sub, err := s.SubscribeForFilter()
	if err != nil {
		return nil, err
	}
	return newFilterPublisher(s.log, sub), nil
}

func (s *publisher) run() {
	defer s.lc.ShutdownCompleted()

loop:
	for {
		select {
		case evt, ok := <-s.parent.Events():
			if !ok {
				s.log.Debugf("parent events closed")
				s.lc.ShutdownInitiated(nil)
				break loop
			}
			s.distributeEvent(evt)
		case resultch := <-s.subscribech:
			resultch <- s.createSubscription()
		case sub := <-s.unsubscribech:
			delete(s.subscriptions, sub)
		}
	}

	for len(s.subscriptions) > 0 {
		s.log.Debugf("draining: %v subscriptions", len(s.subscriptions))
		select {
		case sub := <-s.unsubscribech:
			delete(s.subscriptions, sub)
		}
	}

	<-s.parent.Done()
}

func (s *publisher) distributeEvent(evt Event) {
	s.log.Debugf("distribute event: sending %v to %v subscriptions", evt, len(s.subscriptions))

	for sub := range s.subscriptions {
		sub.send(evt)
	}
}

func (s *publisher) createSubscription() Subscription {
	s.log.Debugf("create subscription: current count %v", len(s.subscriptions))

	sub := newSubscription(s.log, s.lc.ShuttingDown(), s.parent.Ready(), s.parent.Cache())

	s.subscriptions[sub] = struct{}{}

	go func() {
		select {
		case <-sub.Done():
			s.log.Debugf("create subscription: subscription done")
		case <-s.lc.ShuttingDown():
			s.log.Debugf("create subscription: shut down, killing subscription")
			sub.Close()
			<-sub.Done()
		}
		s.unsubscribech <- sub
		s.log.Debugf("create subscription: unsubscribed")
	}()

	return sub
}

func newFilterPublisher(log logutil.Log, subscription FilterSubscription) FilterController {
	return &filterController{subscription, newPublisher(log, subscription)}
}

type filterController struct {
	subscription FilterSubscription
	parent       Controller
}

func (c *filterController) Cache() CacheReader {
	return c.parent.Cache()
}

func (c *filterController) Ready() <-chan struct{} {
	return c.parent.Ready()
}

func (c *filterController) Subscribe() (Subscription, error) {
	return c.parent.Subscribe()
}

func (c *filterController) SubscribeWithFilter(f filter.Filter) (FilterSubscription, error) {
	return c.parent.SubscribeWithFilter(f)
}

func (c *filterController) SubscribeForFilter() (FilterSubscription, error) {
	return c.parent.SubscribeForFilter()
}

func (c *filterController) Clone() (Controller, error) {
	return c.parent.Clone()
}

func (c *filterController) CloneWithFilter(f filter.Filter) (FilterController, error) {
	return c.parent.CloneWithFilter(f)
}

func (c *filterController) CloneForFilter() (FilterController, error) {
	return c.parent.CloneForFilter()
}

func (c *filterController) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *filterController) Close() {
	c.parent.Close()
}

func (c *filterController) Error() error {
	return c.parent.Error()
}

func (c *filterController) Refilter(filter filter.Filter) error {
	return c.subscription.Refilter(filter)
}
