package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/drivers"
	"github.com/rustyeddy/devices/gpio"

	"github.com/rustyeddy/otto/messenger"
	"github.com/rustyeddy/otto/messenger/codec"
	mqttpaho "github.com/rustyeddy/otto/messenger/mqtt"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	broker := getenv("MQTT_BROKER", "tcp://127.0.0.1:1883")
	prefix := getenv("MQTT_PREFIX", "otto")

	slog.Info("starting demo", "broker", broker, "prefix", prefix)

	// ---- MQTT client (Paho) ----
	m := mqttpaho.New(mqttpaho.Config{
		Broker:       broker,
		ClientID:     getenv("MQTT_CLIENT_ID", ""), // auto if empty
		Username:     getenv("MQTT_USER", ""),
		Password:     getenv("MQTT_PASS", ""),
		CleanSession: true,
	})

	reg := messenger.NewRegistry(m, messenger.TopicScheme{Prefix: prefix})
	reg.RetainState = true
	reg.RetainMeta = true
	reg.CommandTimeout = 2 * time.Second

	// important: resubscribe on connect/reconnect
	m.SetOnConnect(func() { reg.ResubscribeAll(ctx) })

	if err := m.Connect(ctx); err != nil {
		slog.Error("mqtt connect failed", "error", err)
		os.Exit(1)
	}
	reg.ResubscribeAll(ctx)

	// ---- Driver factory ----
	// For demo, use virtual GPIO you can run anywhere.
	// Swap to drivers.NewGPIOCDevFactory() on Linux with real pins.
	f := drivers.NewVPIOFactory()

	// ---- Devices ----
	btn := gpio.NewButton(gpio.ButtonConfig{
		Name:     "demo_button",
		Factory:  f,
		Chip:     "vpio",
		Offset:   1,
		Bias:     drivers.BiasPullUp,
		Edge:     drivers.EdgeBoth,
		Debounce: 30 * time.Millisecond,
	})

	rel := gpio.NewRelay(gpio.RelayConfig{
		Name:    "demo_relay",
		Factory: f,
		Chip:    "vpio",
		Offset:  2,
		Initial: false,
	})

	reg.Add(btn)
	reg.Add(rel)

	// ---- Messaging wiring ----
	messenger.WireSource(ctx, reg, btn, codec.JSON[bool]{})
	messenger.WireDuplex(ctx, reg, rel, codec.JSON[bool]{})

	// ---- Demo behavior: press button toggles relay (in OttO, not in devices) ----
	go func() {
		for {
			select {
			case v := <-btn.Out():
				// interpret "true" as press (depends on pull-up/down wiring)
				if v {
					// toggle the relay by sending into its command channel
					rel.In() <- !current(rel)
					slog.Info("button pressed: toggling relay")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// ---- Demo: simulate button presses in VPIO every 3 seconds ----
	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		state := false
		for {
			select {
			case <-t.C:
				state = !state
				edge := drivers.EdgeFalling
				if state {
					edge = drivers.EdgeRising
				}
				f.InjectEdge("vpio", 1, edge, state)
				slog.Info("simulated button edge", "state", state, "edge", edge)
			case <-ctx.Done():
				return
			}
		}
	}()

	// ---- Run ----
	if err := reg.Run(ctx); err != nil {
		slog.Error("registry exited with error", "error", err)
		os.Exit(1)
	}

	slog.Info("demo stopped")
}

// current tries to read the latest state from the relay's Out channel without blocking.
// If nothing is available, it returns false (initial default).
func current(rel *gpio.Relay) bool {
	select {
	case v := <-rel.Out():
		return v
	default:
		return false
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
