package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/devices/gpio" // your devices
	"github.com/rustyeddy/otto/messenger"
	"github.com/rustyeddy/otto/messenger/codec"
	mqttpaho "github.com/rustyeddy/otto/messenger/mqtt"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// MQTT
	m := mqttpaho.New(mqttpaho.Config{
		Broker:       "tcp://10.11.1.11:1883",
		ClientID:     "", // auto
		Username:     "",
		Password:     "",
		CleanSession: true,
	})

	reg := messenger.NewRegistry(m, messenger.TopicScheme{Prefix: "otto"})

	// IMPORTANT: resubscribe on reconnect
	m.SetOnConnect(func() { reg.ResubscribeAll(ctx) })

	if err := m.Connect(ctx); err != nil {
		panic(err)
	}
	// First connect: apply subscriptions now too
	reg.ResubscribeAll(ctx)

	// Devices (examples)
	btn := gpio.NewButton("button") // devices.Source[bool]
	rel := gpio.NewRelay("light")   // devices.Duplex[bool]

	reg.Add(btn)
	reg.Add(rel)

	// Typed wiring
	messenger.WireSource(ctx, reg, btn, codec.JSON[bool]{})
	messenger.WireDuplex(ctx, reg, rel, codec.JSON[bool]{})

	// Run
	_ = reg.Run(ctx)
}
