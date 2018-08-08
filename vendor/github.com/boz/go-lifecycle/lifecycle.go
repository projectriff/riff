package lifecycle

import (
	"context"
	"errors"
)

var ErrRunning = errors.New("lifecycle: still running")

type Lifecycle interface {
	LifecycleReader

	// ShutdownRequest() returns a channel that is available for reading when
	// a shutdown has requested.
	ShutdownRequest() <-chan error

	// ShutdownInitiated() declares that shutdown has begun.  Will panic if called twice.
	ShutdownInitiated(error)

	// ShutdownCompleted() declares that shutdown has completed.  Will panic if called twice.
	ShutdownCompleted()

	// WatchContext() observes the given context and initiates a shutdown
	// if the context is shutdown before the lifecycle is.
	WatchContext(context.Context)

	// Begins shutdown when given channel is ready for reading.
	WatchChannel(<-chan struct{})

	// Shutdown() initiates shutdown by sending a value to the channel
	// requtned by ShutdownRequest() and blocks untill ShutdownCompleted()
	// is called.
	Shutdown(error)

	// Initiate shutdown but does not block until complete.
	ShutdownAsync(error)
}

// LifecycleReader exposes read-only access to lifecycle state.
type LifecycleReader interface {
	// ShuttingDown() returns a channel that is available for reading
	// after ShutdownInitiated() has been called.
	ShuttingDown() <-chan struct{}

	// Done() returns a channel that is available for reading
	// after ShutdownCompleted() has been called.
	Done() <-chan struct{}

	Error() error
}

type lifecycle struct {
	stopch     chan error
	stoppingch chan struct{}
	stoppedch  chan struct{}
	reason     error
}

func New() Lifecycle {
	return &lifecycle{
		stopch:     make(chan error),
		stoppingch: make(chan struct{}),
		stoppedch:  make(chan struct{}),
	}
}

func (l *lifecycle) ShutdownRequest() <-chan error {
	return l.stopch
}

func (l *lifecycle) ShutdownInitiated(err error) {
	l.reason = err
	close(l.stoppingch)
}

func (l *lifecycle) ShuttingDown() <-chan struct{} {
	return l.stoppingch
}

func (l *lifecycle) ShutdownCompleted() {
	close(l.stoppedch)
}

func (l *lifecycle) Done() <-chan struct{} {
	return l.stoppedch
}

func (l *lifecycle) Error() error {
	select {
	case <-l.stoppingch:
		return l.reason
	default:
		return ErrRunning
	}
}

func (l *lifecycle) Shutdown(err error) {
	select {
	case <-l.stoppedch:
		return
	case l.stopch <- err:
	case <-l.stoppingch:
	}
	<-l.stoppedch
}

func (l *lifecycle) ShutdownAsync(err error) {
	select {
	case <-l.stoppedch:
	case <-l.stoppingch:
	case l.stopch <- err:
	}
}

func (l *lifecycle) WatchContext(ctx context.Context) {
	donech := ctx.Done()
	var stopch chan error
	for {
		select {
		case <-l.stoppingch:
			return
		case <-donech:
			donech = nil
			stopch = l.stopch
		case stopch <- ctx.Err():
			return
		}
	}
}

func (l *lifecycle) WatchChannel(donech <-chan struct{}) {
	var stopch chan error
	for {
		select {
		case <-l.stoppingch:
			return
		case <-donech:
			donech = nil
			stopch = l.stopch
		case stopch <- nil:
			return
		}
	}
}
