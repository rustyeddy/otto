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

func NewFollow[T any](name string, src devices.Source[T], dst devices.Sink[T]) *Follow[T] {
	return &Follow[T]{name: name, Src: src, Dst: dst}
}

func (f *Follow[T]) Name() string { return f.name }

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
