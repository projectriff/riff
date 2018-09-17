package kcache

import (
	lifecycle "github.com/boz/go-lifecycle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Monitor interface {
	Close()
	Done() <-chan struct{}
	Error() error
}

type Handler interface {
	OnInitialize([]metav1.Object)
	OnCreate(metav1.Object)
	OnUpdate(metav1.Object)
	OnDelete(metav1.Object)
}

type HandlerBuilder interface {
	OnInitialize(func([]metav1.Object)) HandlerBuilder
	OnCreate(func(metav1.Object)) HandlerBuilder
	OnUpdate(func(metav1.Object)) HandlerBuilder
	OnDelete(func(metav1.Object)) HandlerBuilder
	Create() Handler
}

func BuildHandler() HandlerBuilder {
	return &handlerBuilder{}
}

type handler struct {
	onInitialize func([]metav1.Object)
	onCreate     func(metav1.Object)
	onUpdate     func(metav1.Object)
	onDelete     func(metav1.Object)
}

type handlerBuilder handler

func (hb *handlerBuilder) OnInitialize(fn func([]metav1.Object)) HandlerBuilder {
	hb.onInitialize = fn
	return hb
}

func (hb *handlerBuilder) OnCreate(fn func(metav1.Object)) HandlerBuilder {
	hb.onCreate = fn
	return hb
}

func (hb *handlerBuilder) OnUpdate(fn func(metav1.Object)) HandlerBuilder {
	hb.onUpdate = fn
	return hb
}

func (hb *handlerBuilder) OnDelete(fn func(metav1.Object)) HandlerBuilder {
	hb.onDelete = fn
	return hb
}

func (hb *handlerBuilder) Create() Handler {
	return handler(*hb)
}

func (h handler) OnInitialize(objs []metav1.Object) {
	if h.onInitialize != nil {
		h.onInitialize(objs)
	}
}

func (h handler) OnCreate(obj metav1.Object) {
	if h.onCreate != nil {
		h.onCreate(obj)
	}
}

func (h handler) OnUpdate(obj metav1.Object) {
	if h.onUpdate != nil {
		h.onUpdate(obj)
	}
}

func (h handler) OnDelete(obj metav1.Object) {
	if h.onDelete != nil {
		h.onDelete(obj)
	}
}

func NewMonitor(publisher Publisher, handler Handler) (Monitor, error) {
	sub, err := publisher.Subscribe()
	if err != nil {
		return nil, err
	}
	m := &monitor{sub, handler, lifecycle.New()}
	go m.run()
	return m, nil
}

type monitor struct {
	sub     Subscription
	handler Handler
	lc      lifecycle.Lifecycle
}

func (m *monitor) run() {
	defer m.lc.ShutdownCompleted()

	select {
	case <-m.sub.Done():
		m.lc.ShutdownInitiated(nil)
		return
	case <-m.sub.Ready():
		objs, err := m.sub.Cache().List()
		if err != nil {
			m.lc.ShutdownInitiated(err)
			m.sub.Close()
			<-m.sub.Done()
			return
		}
		m.handler.OnInitialize(objs)
	}

	for {
		select {
		case <-m.sub.Done():
			m.lc.ShutdownInitiated(nil)
			return
		case ev, ok := <-m.sub.Events():
			if !ok {
				m.lc.ShutdownInitiated(nil)
				<-m.sub.Done()
				return
			}
			switch ev.Type() {
			case EventTypeCreate:
				m.handler.OnCreate(ev.Resource())
			case EventTypeUpdate:
				m.handler.OnUpdate(ev.Resource())
			case EventTypeDelete:
				m.handler.OnDelete(ev.Resource())
			}
		}
	}

}

func (m *monitor) Close() {
	m.sub.Close()
}

func (m *monitor) Done() <-chan struct{} {
	return m.lc.Done()
}

func (m *monitor) Error() error {
	if err := m.lc.Error(); err != nil {
		return err
	}
	return m.sub.Error()
}
