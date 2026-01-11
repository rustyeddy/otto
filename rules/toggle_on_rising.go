package rules

import (
	"context"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/otto/messenger"
)

type ToggleOnRisingEdge struct {
	name string

	Button devices.Source[bool]
	Relay  devices.Duplex[bool]

	Registry *messenger.Registry

	// interpret press as true or false (depends on wiring/pull-up)
	PressValue bool

	// simple guard against rapid repeats
	MinInterval time.Duration
}

// NewToggleOnRisingEdge returns a rule that toggles a relay on rising edge presses.
func NewToggleOnRisingEdge(name string, reg *messenger.Registry, btn devices.Source[bool], relay devices.Duplex[bool]) *ToggleOnRisingEdge {
	return &ToggleOnRisingEdge{
		name:        name,
		Button:      btn,
		Relay:       relay,
		Registry:    reg,
		PressValue:  true,
		MinInterval: 150 * time.Millisecond,
	}
}

// Name returns the rule name.
func (t *ToggleOnRisingEdge) Name() string { return t.name }

// Run listens for button presses and toggles the relay.
func (t *ToggleOnRisingEdge) Run(ctx context.Context) error {

	var last time.Time

	for {
		select {
		case v, ok := <-t.Button.Out():
			if !ok {
				return nil
			}
			if v != t.PressValue {
				continue
			}
			now := time.Now()
			if t.MinInterval > 0 && !last.IsZero() && now.Sub(last) < t.MinInterval {
				continue
			}
			last = now

			// read cached relay state
			cur, ok := messenger.StateAs[bool](t.Registry, t.Relay.Name())
			if !ok {
				cur = false // default
			}

			// Toggle
			select {
			case t.Relay.In() <- !cur:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
