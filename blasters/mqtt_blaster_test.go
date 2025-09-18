package blasters

import (
	"testing"
	"time"
)

func TestNewMQTTBlasters(t *testing.T) {
	mb := NewMQTTBlasters(3)
	if mb.Count != 3 {
		t.Errorf("expected Count=3, got %d", mb.Count)
	}
	if len(mb.Blasters) != 3 {
		t.Errorf("expected 3 Blasters, got %d", len(mb.Blasters))
	}
	for i, b := range mb.Blasters {
		if b == nil {
			t.Errorf("Blaster at index %d is nil", i)
		}
	}
}

func TestNewMQTTBlasters_Zero(t *testing.T) {
	mb := NewMQTTBlasters(0)
	if mb.Count != 0 {
		t.Errorf("expected Count=0, got %d", mb.Count)
	}
	if len(mb.Blasters) != 0 {
		t.Errorf("expected 0 Blasters, got %d", len(mb.Blasters))
	}
}

func TestMQTTBlasters_Stop(t *testing.T) {
	mb := NewMQTTBlasters(1)
	mb.Running = true
	mb.Stop()
	if mb.Running {
		t.Error("Stop() did not set Running to false")
	}
}

// This test will run Blast in a goroutine and stop it after a short delay.
// It does not test actual MQTT publishing, just the control flow.
func TestMQTTBlasters_Blast(t *testing.T) {
	mb := NewMQTTBlasters(2)
	mb.Wait = 10 // 10ms for fast test
	ch := make(chan struct{})
	go func() {
		_ = mb.Blast()
		ch <- struct{}{}
	}()
	time.Sleep(30 * time.Millisecond)
	mb.Stop()
	select {
	case <-ch:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("Blast did not exit after Stop() was called")
	}
}

func TestMQTTBlasters_Blast_NoBlasters(t *testing.T) {
	mb := NewMQTTBlasters(0)
	mb.Wait = 1
	ch := make(chan struct{})
	go func() {
		_ = mb.Blast()
		ch <- struct{}{}
	}()
	time.Sleep(5 * time.Millisecond)
	mb.Stop()
	select {
	case <-ch:
		// success
	case <-time.After(50 * time.Millisecond):
		t.Error("Blast did not exit with zero blasters")
	}
}
