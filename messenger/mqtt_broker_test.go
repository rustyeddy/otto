package messenger

import (
	"context"
	"strings"
	"testing"
	"time"
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
	if err := StopMQTTBroker(context.Background()); err != nil {
		t.Fatalf("expected nil error from StopMQTTBroker when no broker, got: %v", err)
	}
}

func TestStartMQTTBrokerStartAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownFn, err := StartMQTTBroker(ctx)
	if err != nil {
		if isAddrInUse(err) {
			t.Skipf("mqtt broker port in use, skipping test: %v", err)
		}
		t.Fatalf("unexpected error starting mqtt broker: %v", err)
	}
	if shutdownFn == nil {
		t.Fatalf("expected non-nil shutdown function")
	}

	// Give broker a moment to start
	time.Sleep(150 * time.Millisecond)

	// Call package-level stop helper and the returned shutdown to ensure graceful shutdown.
	if err := StopMQTTBroker(context.Background()); err != nil {
		t.Fatalf("StopMQTTBroker returned error: %v", err)
	}
	// Also explicitly call returned shutdown function to ensure it's callable.
	if err := shutdownFn(context.Background()); err != nil {
		t.Fatalf("shutdown function returned error: %v", err)
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
		t.Fatalf("unexpected error starting mqtt broker: %v", err)
	}
	if shutdownFn == nil {
		cancel()
		t.Fatalf("expected non-nil shutdown function")
	}

	// Cancel the context and allow the broker's goroutine to trigger shutdown.
	cancel()
	time.Sleep(200 * time.Millisecond)

	// Ensure calling StopMQTTBroker after context cancellation is safe.
	if err := StopMQTTBroker(context.Background()); err != nil {
		t.Fatalf("StopMQTTBroker returned error after context cancel: %v", err)
	}

	// Defensive call to returned shutdown function (should be safe even if already closed).
	_ = shutdownFn(context.Background())
	time.Sleep(100 * time.Millisecond)
}
