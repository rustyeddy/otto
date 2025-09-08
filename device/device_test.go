package device

import (
	"fmt"
	"testing"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

func TestMock(t *testing.T) {
	mocks := []bool{true, false}

	if mock != false {
		t.Error("Expected mock (false) got (true)")
	}

	for _, b := range mocks {
		Mock(b)
		if IsMock() != b {
			t.Errorf("Expected mock (%t) got (%t)", b, mock)
		}
	}
}

func TestNewDevice(t *testing.T) {
	name := "test-device"
	dev := NewDevice(name)
	if dev.Name() != name {
		t.Errorf("Expected test device name (%s) got (%s)", name, dev.Name())
	}

	str := dev.String()
	if str == "" {
		t.Errorf("Expected device string to not be null but it was")
	}

	jbytes, err := dev.JSON()
	if err != nil {
		t.Error("Error creating JSON from device: ", err)
	}

	if len(jbytes) == 0 {
		t.Errorf("Expected JSON bytes but got (0)")
	}
}

func TestDeviceString(t *testing.T) {
	println("test-device-string")
}
func TestTimerLoop(t *testing.T) {
	done := make(chan any)
	readpubCalled := false

	readpub := func() error {
		readpubCalled = true
		return nil
	}

	device := NewDevice("test-device")
	period := 100 * time.Millisecond

	// Start TimerLoop in a goroutine
	go device.TimerLoop(period, done, readpub)

	// Wait for the ticker to trigger at least once
	time.Sleep(2 * period)

	// Verify readpub was called
	if !readpubCalled {
		t.Error("Expected readpub to be called, but it was not")
	}

	// Stop the TimerLoop
	done <- struct{}{}

	// Allow some time for the goroutine to exit
	time.Sleep(100 * time.Millisecond)
}

func TestTimerLoopWithError(t *testing.T) {
	done := make(chan any)
	readpubCalled := false

	readpub := func() error {
		readpubCalled = true
		return fmt.Errorf("mock error")
	}

	device := NewDevice("test-device")
	period := 100 * time.Millisecond

	// Start TimerLoop in a goroutine
	go device.TimerLoop(period, done, readpub)

	// Wait for the ticker to trigger at least once
	time.Sleep(2 * period)

	// Verify readpub was called
	if !readpubCalled {
		t.Error("Expected readpub to be called, but it was not")
	}

	// Stop the TimerLoop
	done <- struct{}{}

	// Allow some time for the goroutine to exit
	time.Sleep(100 * time.Millisecond)
}

func TestTimerLoopNoPeriod(t *testing.T) {
	done := make(chan any)
	readpub := func() error {
		t.Error("readpub should not be called when period is 0")
		return nil
	}

	device := NewDevice("test-device")
	period := time.Duration(0)

	// Start TimerLoop
	device.TimerLoop(period, done, readpub)

	// Verify TimerLoop exits immediately
	select {
	case <-done:
		t.Error("done channel should not be used when period is 0")
	default:
	}
}
func TestDeviceMessangerInitialization(t *testing.T) {
	name := "test-device"
	dev := NewDevice(name)

	if dev.Messanger == nil {
		t.Error("Expected Messanger to be initialized, but it was nil")
	}

	if dev.Messanger.ID() != name {
		t.Errorf("Expected Messanger name to be '%s', but got '%s'", name, dev.Messanger.ID())
	}
}

func TestDeviceMessangerPublish(t *testing.T) {
	name := "test-device"
	dev := NewDevice(name)

	topic := "test/topic"
	message := "test message"

	dev.Messanger.SetTopic(topic)
	dev.Messanger.Publish(message)
	err := dev.Messanger.Error()
	if err != nil {
		t.Errorf("Expected Publish to succeed, but got error: %v", err)
	}
}

func TestDeviceMessangerSubscribe(t *testing.T) {
	name := "test-device"
	dev := NewDevice(name)

	topic := "dev/" + name
	msgstr := "test message"
	messageReceived := false
	m := messanger.New(topic, []byte(msgstr), "test")

	dev.Messanger.Subscribe(topic, func(msg *messanger.Msg) {
		if msg.String() == m.String() {
			messageReceived = true
		}
	})
	err := dev.Messanger.Error()
	if err != nil {
		t.Errorf("Expected Subscribe to succeed, but got error: %v", err)
	}

	// Simulate publishing a message to the topic
	dev.Messanger.PubMsg(m)
	err = dev.Messanger.Error()
	if err != nil {
		t.Errorf("Expected Publish to succeed, but got error: %v", err)
	}

	// Allow some time for the message to be received
	time.Sleep(100 * time.Millisecond)

	if !messageReceived {
		t.Error("Expected to receive the published message, but did not")
	}
}
