package kcache

import (
	"math/rand"
	"time"
)

type ticker interface {
	Next() <-chan int
	Reset()
	Stop()
	Done() <-chan struct{}
}

func newTicker(period time.Duration, fuzz float64) ticker {

	t := &_ticker{
		period:  period,
		fuzz:    fuzz,
		nextch:  make(chan int),
		resetch: make(chan bool),
		stopch:  make(chan bool),
		donech:  make(chan struct{}),
	}

	go t.run()

	return t
}

type _ticker struct {
	period time.Duration
	fuzz   float64

	nextch  chan int
	resetch chan bool
	stopch  chan bool
	donech  chan struct{}
}

func (t *_ticker) Next() <-chan int {
	return t.nextch
}

func (t *_ticker) Reset() {
	select {
	case t.resetch <- true:
	case <-t.donech:
	}
}

func (t *_ticker) Stop() {
	select {
	case t.stopch <- true:
	case <-t.donech:
	}
}

func (t *_ticker) Done() <-chan struct{} {
	return t.donech
}

func (t *_ticker) run() {
	defer close(t.donech)

	count := 0
	timer := time.NewTimer(t.nextPeriod())

	var nextch chan int

	for {

		select {

		case <-t.resetch:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(t.nextPeriod())
			nextch = nil

		case <-t.stopch:
			timer.Stop()
			return

		case <-timer.C:
			timer.Stop()
			nextch = t.nextch

		case nextch <- count:
			count++
			nextch = nil
			timer.Reset(t.nextPeriod())

		}
	}
}

func (t *_ticker) nextPeriod() time.Duration {
	delta := t.fuzz * float64(t.period)

	min := float64(t.period) - delta
	max := float64(t.period) + delta

	r := rand.Float64()

	return time.Duration(min + r*(max-min+1))

}
