package messanger

import (
	"context"
	"log"

	mqttserver "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// StartBroker starts an embedded MQTT broker and returns a shutdown func.
// Minimal config: in-memory (no persistence), TCP listener on :1883.
func StartMQTTBroker(ctx context.Context) (func(context.Context) error, error) {
	// Create broker with default options (in-memory state).
	srv := mqttserver.New(nil)

	// TCP listener (native MQTT)
	cfg := listeners.Config{
		ID:      "tcp-1883",
		Address: ":1883",
	}
	tcp := listeners.NewTCP(cfg)
	if err := srv.AddListener(tcp); err != nil {
		return nil, err
	}

	// OPTIONAL: WebSocket listener (for browsers / MQTT over WS)
	// ws := listeners.NewWebsocket("ws-1882", ":1882", nil)
	// if err := srv.AddListener(ws); err != nil {
	// 	return nil, err
	// }

	// OPTIONAL: HTTP healthcheck (e.g., for k8s liveness/readiness)
	// hc := listeners.NewHTTPHealthCheck("healthz", ":8081", nil)
	// if err := srv.AddListener(hc); err != nil {
	// 	return nil, err
	// }

	// Run the broker in the background.
	go func() {
		if err := srv.Serve(); err != nil {
			// Serve only returns on fatal error or Close()
			log.Printf("mqtt broker stopped: %v", err)
		}
	}()

	// Close on context cancellation if provided.
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	// Return a shutdown function you can call from your app.
	shutdown := func(_ context.Context) error {
		return srv.Close()
	}
	return shutdown, nil
}
