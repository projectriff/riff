package kcache

import (
	"context"
	"fmt"
	"time"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/boz/kcache/filter"
)

type Builder interface {
	Context(context.Context) Builder
	Log(logutil.Log) Builder

	Filter(filter.Filter) Builder

	Client(client.Client) Builder
	Lister() ListerBuilder
	Watcher() WatcherBuilder

	Create() (Controller, error)
}

type ListerBuilder interface {
	RefreshPeriod(time.Duration) ListerBuilder
	Client(client.ListClient) ListerBuilder
}

type WatcherBuilder interface {
	Client(client.WatchClient) WatcherBuilder
}

func NewBuilder() Builder {
	return &builder{
		filter: filter.Null(),
		log:    logutil.Default(),
		ctx:    context.Background(),
		lb:     newListerBuilder(),
		wb:     newWatcherBuilder(),
	}
}

type builder struct {
	client client.Client
	log    logutil.Log
	ctx    context.Context
	filter filter.Filter

	lb *listerBuilder
	wb *watcherBuilder
}

func (b *builder) Context(ctx context.Context) Builder {
	b.ctx = ctx
	return b
}

func (b *builder) Log(log logutil.Log) Builder {
	b.log = log
	return b
}

func (b *builder) Filter(filter filter.Filter) Builder {
	b.filter = filter
	return b
}

func (b *builder) Client(client client.Client) Builder {
	b.lb.Client(client)
	b.wb.Client(client)
	return b
}

func (b *builder) Lister() ListerBuilder {
	return b.lb
}

func (b *builder) Watcher() WatcherBuilder {
	return b.wb
}

func (b *builder) Create() (Controller, error) {
	if b.log == nil {
		return nil, fmt.Errorf("kcache builder: log required")
	}

	log := b.log.WithComponent("controller")
	ctx := b.ctx

	lc := lifecycle.New()

	cache := newCache(ctx, log, lc.ShuttingDown(), b.filter)
	readych := make(chan struct{})

	subscription := newSubscription(log, lc.ShuttingDown(), readych, cache)
	publisher := newPublisher(log, subscription)

	c := &controller{
		readych: readych,

		subscription: subscription,
		publisher:    publisher,

		lister:  newLister(ctx, log, lc.ShuttingDown(), b.lb.period, b.lb.client),
		watcher: newWatcher(ctx, log, lc.ShuttingDown(), b.wb.client),

		cache: cache,

		log: log,
		lc:  lc,
		ctx: ctx,
	}

	go c.lc.WatchContext(c.ctx)

	go c.run()

	return c, nil
}

type listerBuilder struct {
	client client.ListClient
	period time.Duration
}

func newListerBuilder() *listerBuilder {
	return &listerBuilder{period: defaultRefreshPeriod}
}

func (b *listerBuilder) RefreshPeriod(period time.Duration) ListerBuilder {
	b.period = period
	return b
}

func (b *listerBuilder) Client(client client.ListClient) ListerBuilder {
	b.client = client
	return b
}

type watcherBuilder struct {
	client client.WatchClient
}

func newWatcherBuilder() *watcherBuilder {
	return &watcherBuilder{}
}

func (b *watcherBuilder) Client(client client.WatchClient) WatcherBuilder {
	b.client = client
	return b
}
