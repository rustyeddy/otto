package messenger

import (
	"context"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/otto/messenger/codec"
)

// WireSource publishes device.Out() to MQTT .../state (JSON-encoded).
func WireSource[T any](ctx context.Context, r *Registry, dev devices.Source[T], c codec.Codec[T]) {
	name := dev.Name()

	go func() {
		for {
			select {
			case v, ok := <-dev.Out():
				if !ok {
					return
				}

				// encode
				b, err := c.Marshal(v)
				if err != nil {
					r.Log.Warn("state marshal failed", "device", name, "error", err)
					continue
				}

				// cache
				r.stateMu.Lock()
				r.stateRaw[name] = b
				r.stateAny[name] = v
				r.stateMu.Unlock()

				// publish
				_ = r.MQTT.Publish(ctx, r.Topics.State(name), b, r.RetainState, r.QoSState)

			case <-ctx.Done():
				return
			}
		}
	}()
}

// WireSink subscribes to MQTT .../set and delivers decoded values into device.In().
// Uses timeout so MQTT callback doesn't block forever.
func WireSink[T any](ctx context.Context, r *Registry, dev devices.Sink[T], c codec.Codec[T]) {
	name := dev.Name()
	setTopic := r.Topics.Set(name)
	in := dev.In()

	r.WantSub(setTopic, r.QoSSet, func(m Message) {
		v, err := c.Unmarshal(m.Payload)
		if err != nil {
			r.Log.Warn("set unmarshal failed", "device", name, "topic", m.Topic, "error", err)
			return
		}

		timeout := r.CommandTimeout
		if timeout <= 0 {
			timeout = 2 * time.Second
		}

		select {
		case in <- v:
			// delivered
		case <-time.After(timeout):
			r.Log.Warn("set delivery timeout", "device", name, "topic", m.Topic)
		case <-ctx.Done():
			return
		}
	})
}

// WireDuplex wires both state publish and set subscribe.
func WireDuplex[T any](ctx context.Context, r *Registry, dev devices.Duplex[T], c codec.Codec[T]) {
	WireSource(ctx, r, dev, c)
	WireSink(ctx, r, dev, c)
}
