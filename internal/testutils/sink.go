package testutils

import (
	"context"

	"github.com/rustyeddy/devices"
)

type Sink[T any] struct {
	devices.Base
	in chan T
}

// NewSink creates a devices.Sink[T] fake.
// buf <= 0 defaults to 16.
func NewSink[T any](name string, buf int) *Sink[T] {
	if buf <= 0 {
		buf = 16
	}
	return &Sink[T]{
		Base: devices.NewBase(name, 16),
		in:   make(chan T, buf),
	}
}

func (s *Sink[T]) In() chan<- T  { return s.in }
func (s *Sink[T]) Get() <-chan T { return s.in }

// Read is a convenience for tests: read one value (blocking).
func (s *Sink[T]) Read() T {
	return <-s.in
}

// TryRead is a convenience for tests: non-blocking read.
func (s *Sink[T]) TryRead() (T, bool) {
	select {
	case v := <-s.in:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func (s *Sink[T]) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Note: sinks generally should NOT close their input channel;
// senders own it. So Close() just closes the device.
func (s *Sink[T]) Close() error { return s.Base.Close() }
