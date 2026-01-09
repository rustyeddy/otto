package messenger

import "context"

// Message is a decoded MQTT message delivered to a handler.
type Message struct {
	Topic   string
	Payload []byte
	Retain  bool
	QoS     byte
}

// MQTT abstracts the MQTT client operations used by the messenger.
type MQTT interface {
	// Publish should be safe to call from multiple goroutines.
	Publish(ctx context.Context, topic string, payload []byte, retain bool, qos byte) error
	Subscribe(ctx context.Context, topic string, qos byte, handler func(Message)) (unsubscribe func() error, err error)
	SetWill(topic string, payload []byte, retain bool, qos byte) error
}
