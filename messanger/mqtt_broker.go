package messanger

import (
	"context"
	"log"

	mqttserver "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// StartMQTTBroker starts an embedded MQTT broker for local or embedded deployments.
// This eliminates the need for an external broker like Mosquitto or EMQX.
//
// The broker is configured with:
//   - In-memory state (no persistence to disk)
//   - TCP listener on port 1883 (standard MQTT)
//   - Authentication with two users: "otto"/"otto123" and "admin"/"admin"
//   - Clean session support
//   - Background operation (non-blocking)
//
// The broker automatically shuts down when:
//  1. The provided context is cancelled/done
//  2. The returned shutdown function is called
//
// This is useful for:
//   - Single-device IoT applications
//   - Testing without external dependencies
//   - Development environments
//   - Embedded systems where an external broker isn't practical
//
// Parameters:
//   - ctx: Context for managing broker lifecycle; broker stops when ctx.Done()
//
// Returns:
//   - A shutdown function that can be called to stop the broker gracefully
//   - An error if broker initialization fails (e.g., port already in use)
//
// Example:
//
//	ctx := context.Background()
//	shutdown, err := StartMQTTBroker(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(context.Background())
//
//	// Now create MQTT clients that connect to localhost:1883
//	client := NewMessangerMQTT("client1", "localhost")
//	client.Connect()
//
// Optional Features (commented out):
//   - WebSocket listener on port 1882 for browser clients
//   - HTTP health check endpoint on port 8081 for monitoring
var (
	shutdown func(context.Context) error
)

func StartMQTTBroker(ctx context.Context) (func(context.Context) error, error) {
	// Create broker with default options (in-memory state).
	srv := mqttserver.New(nil)

	// Add authentication hook with username/password
	err := srv.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{
				{
					Username: "otto",
					Password: "otto123",
					Allow:    true,
				},
				{
					Username: "admin",
					Password: "admin",
					Allow:    true,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

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
	shutdown = func(_ context.Context) error {
		return srv.Close()
	}
	return shutdown, nil
}

func StopMQTTBroker(ctx context.Context) error {
	if shutdown != nil {
		return shutdown(ctx)
	}
	return nil
}
