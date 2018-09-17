package kcache

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/pkg/errors"
)

var (
	errInvalidType = fmt.Errorf("Invalid type")
)

const (
	defaultRefreshPeriod = time.Minute
	defaultRefreshFuzz   = 0.10
)

type lister interface {
	Result() <-chan listResult
	Done() <-chan struct{}
	Error() error
}

type listResult struct {
	list runtime.Object
	err  error
}

type _lister struct {
	client   client.ListClient
	period   time.Duration
	resultch chan listResult

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func newLister(ctx context.Context, log logutil.Log, stopch <-chan struct{}, period time.Duration, client client.ListClient) *_lister {
	log = log.WithComponent("lister")

	l := &_lister{
		client:   client,
		period:   period,
		resultch: make(chan listResult),
		log:      log,
		lc:       lifecycle.New(),
		ctx:      ctx,
	}

	go l.lc.WatchContext(ctx)
	go l.lc.WatchChannel(stopch)

	go l.run()

	return l
}

func (l *_lister) Result() <-chan listResult {
	return l.resultch
}

func (l *_lister) Done() <-chan struct{} {
	return l.lc.Done()
}

func (l *_lister) Error() error {
	return l.lc.Error()
}

func (l *_lister) run() {
	defer l.lc.ShutdownCompleted()

	var resultch chan listResult
	var result listResult

	runch, donech := l.list()

	ticker := newTicker(l.period, defaultRefreshFuzz)
	var tickch <-chan int

mainloop:
	for {
		select {
		case <-tickch:
			runch, donech = l.list()
			tickch = nil

		case result = <-runch:
			resultch = l.resultch
			runch = nil

		case resultch <- result:
			ticker.Reset()
			resultch = nil
			tickch = ticker.Next()

		case err := <-l.lc.ShutdownRequest():
			l.lc.ShutdownInitiated(err)
			break mainloop
		}
	}

	ticker.Stop()
	<-ticker.Done()
	<-donech
}

func (l *_lister) list() (<-chan listResult, <-chan struct{}) {
	runch := make(chan listResult, 1)
	donech := make(chan struct{})
	ctx, cancel := context.WithCancel(l.ctx)

	go func() {
		defer cancel()
		select {
		case <-l.lc.ShuttingDown():
		case <-donech:
		}
	}()

	go func() {
		defer close(donech)
		runch <- l.executeList(ctx)
	}()

	return runch, donech
}

func (l *_lister) executeList(ctx context.Context) listResult {
	list, err := l.client.List(ctx, v1.ListOptions{})

	if err != nil {
		if err != context.Canceled {
			l.log.Errorf("client list: %v", err)
		}
		return listResult{nil, errors.Wrap(err, "client list")}
	}

	if _, ok := list.(meta.List); !ok {
		l.log.Errorf("invalid type: %T", list)
		return listResult{nil, errors.WithStack(errInvalidType)}
	}

	return listResult{list, nil}
}
