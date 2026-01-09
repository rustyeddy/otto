package messenger

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func isAddrInUse(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "address already in use") || strings.Contains(s, "bind:") || strings.Contains(s, "bind ") {
		return true
	}
	return false
}

func TestStopMQTTBrokerNoop(t *testing.T) {
	// Ensure calling StopMQTTBroker before starting broker is a no-op and doesn't panic.
	err := StopMQTTBroker(context.Background())
	assert.NoError(t, err)
}

func TestStartMQTTBrokerStartAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownFn, err := StartMQTTBroker(ctx)
	if err != nil {
		if isAddrInUse(err) {
			t.Skipf("mqtt broker port in use, skipping test: %v", err)
		}
		if !assert.NoError(t, err) {
			return
		}
	}
	if !assert.NotNil(t, shutdownFn) {
		return
	}

	// Give broker a moment to start
	time.Sleep(150 * time.Millisecond)

	// Call package-level stop helper and the returned shutdown to ensure graceful shutdown.
	if !assert.NoError(t, StopMQTTBroker(context.Background())) {
		return
	}
	// Also explicitly call returned shutdown function to ensure it's callable.
	if !assert.NoError(t, shutdownFn(context.Background())) {
		return
	}

	// Allow shutdown to complete
	time.Sleep(150 * time.Millisecond)
}

func TestStartMQTTBrokerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	shutdownFn, err := StartMQTTBroker(ctx)
	if err != nil {
		cancel()
		if isAddrInUse(err) {
			t.Skipf("mqtt broker port in use, skipping test: %v", err)
		}
		if !assert.NoError(t, err) {
			return
		}
	}
	if !assert.NotNil(t, shutdownFn) {
		cancel()
		return
	}

	// Cancel the context and allow the broker's goroutine to trigger shutdown.
	cancel()
	time.Sleep(200 * time.Millisecond)

	// Ensure calling StopMQTTBroker after context cancellation is safe.
	if !assert.NoError(t, StopMQTTBroker(context.Background())) {
		return
	}

	// Defensive call to returned shutdown function (should be safe even if already closed).
	_ = shutdownFn(context.Background())
	time.Sleep(100 * time.Millisecond)
}
