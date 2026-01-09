package mqtt

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/rustyeddy/otto/messenger"
)

type Paho struct {
	opts *paho.ClientOptions
	c    paho.Client

	// Called whenever Paho connects/reconnects.
	onConnect func()
}

type Config struct {
	Broker   string // e.g. "tcp://10.11.0.10:1883"
	ClientID string // if empty, random
	Username string
	Password string

	CleanSession bool
}

func New(cfg Config) *Paho {
	id := cfg.ClientID
	if id == "" {
		id = "otto-" + randSuffix()
	}

	opts := paho.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(id).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second).
		SetCleanSession(cfg.CleanSession)

	p := &Paho{opts: opts}

	opts.SetConnectionLostHandler(func(_ paho.Client, err error) {
		slog.Info("MQTT disconnected", "error", err)
	})

	opts.OnConnect = func(_ paho.Client) {
		slog.Info("MQTT connected")
		if p.onConnect != nil {
			p.onConnect()
		}
	}

	return p
}

func randSuffix() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (p *Paho) SetOnConnect(fn func()) {
	p.onConnect = fn
}

func (p *Paho) Connect(ctx context.Context) error {
	if p.c == nil {
		p.c = paho.NewClient(p.opts)
	}
	tok := p.c.Connect()
	if !tok.WaitTimeout(15 * time.Second) {
		return errors.New("mqtt connect timeout")
	}
	return tok.Error()
}

func (p *Paho) SetWill(topic string, payload []byte, retain bool, qos byte) error {
	if p.opts == nil {
		return errors.New("mqtt options not initialized")
	}
	// Paho expects string payload for will.
	p.opts.SetWill(topic, string(payload), qos, retain)
	return nil
}

func (p *Paho) Publish(ctx context.Context, topic string, payload []byte, retain bool, qos byte) error {
	tok := p.c.Publish(topic, qos, retain, payload)
	// For QoS0, we usually don't need to wait. For QoS1, wait briefly.
	if qos > 0 {
		if !tok.WaitTimeout(5 * time.Second) {
			return errors.New("mqtt publish timeout")
		}
	}
	return tok.Error()
}

func (p *Paho) Subscribe(ctx context.Context, topic string, qos byte, handler func(messenger.Message)) (func() error, error) {
	tok := p.c.Subscribe(topic, qos, func(_ paho.Client, msg paho.Message) {
		handler(messenger.Message{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
			Retain:  msg.Retained(),
			QoS:     msg.Qos(),
		})
	})
	if !tok.WaitTimeout(10 * time.Second) {
		return nil, errors.New("mqtt subscribe timeout")
	}
	if tok.Error() != nil {
		return nil, tok.Error()
	}

	return func() error {
		ut := p.c.Unsubscribe(topic)
		if !ut.WaitTimeout(10 * time.Second) {
			return errors.New("mqtt unsubscribe timeout")
		}
		return ut.Error()
	}, nil
}
