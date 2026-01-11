package rules

import (
	"context"

	"github.com/rustyeddy/devices"
)

type Follow[T any] struct {
	name string
	Src  devices.Source[T]
	Dst  devices.Sink[T]
}

// NewFollow returns a rule that copies source values into the sink.
func NewFollow[T any](name string, src devices.Source[T], dst devices.Sink[T]) *Follow[T] {
	return &Follow[T]{name: name, Src: src, Dst: dst}
}

// Name returns the rule name.
func (f *Follow[T]) Name() string { return f.name }

// Run forwards source values into the sink until context cancellation.
func (f *Follow[T]) Run(ctx context.Context) error {
	for {
		select {
		case v, ok := <-f.Src.Out():
			if !ok {
				return nil
			}
			select {
			case f.Dst.In() <- v:
			case <-ctx.Done():
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}
