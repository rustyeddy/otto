package testutils

import (
	"context"
	"sync"

	"github.com/rustyeddy/devices"
)

type Source[T any] struct {
	devices.Base
	out      chan T
	closeOut sync.Once
}

// NewSource creates a devices.Source[T] fake.
// buf <= 0 defaults to 16.
func NewSource[T any](name string, buf int) *Source[T] {
	if buf <= 0 {
		buf = 16
	}
	return &Source[T]{
		Base: devices.NewBase(name, 16),
		out:  make(chan T, buf),
	}
}

func (s *Source[T]) Set() chan<- T { return s.out }
func (s *Source[T]) Out() <-chan T { return s.out }

// Emit pushes a value into the source.
// Handy for tests that want to drive the stream.
func (s *Source[T]) Emit(v T) {
	s.out <- v
}

// CloseOut closes the Out() channel exactly once.
// Useful if a test wants to end the stream deterministically.
func (s *Source[T]) CloseOut() {
	s.closeOut.Do(func() { close(s.out) })
}

func (s *Source[T]) Run(ctx context.Context) error {
	<-ctx.Done()
	s.CloseOut()
	return ctx.Err()
}

func (s *Source[T]) Close() error {
	// Don’t auto-close out here unless you want Close() to end streams.
	// I usually *do* close it, because tests often won’t call Run().
	s.CloseOut()
	return s.Base.Close()
}
