package otto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOttOInit(t *testing.T) {
	o := &OttO{
		Name:       "TestOttO",
		MQTTBroker: "localhost",
	}

	o.Init()

	assert.NotNil(t, o.done, "Expected done channel to be initialized")
	assert.NotNil(t, o.Messenger, "Expected Messanger to be initialized")
	assert.NotNil(t, o.StationManager, "Expected StationManager to be initialized")
	assert.NotNil(t, o.Station, "Expected Station to be initialized")
	assert.NotNil(t, o.Server, "Expected Server to be initialized")
	// Clean up by stopping OttO (handle done channel in goroutine)
	go func() {
		<-o.Done()
	}()
	o.Stop()
}

func TestOttOBrokerShutdown(t *testing.T) {
	o := &OttO{Name: "TestOttOShutdown"}
	o.Init()

	// Test that calling Stop doesn't panic and properly shuts down the broker
	assert.NotPanics(t, func() {
		// Start a goroutine to receive from done channel
		go func() {
			<-o.Done()
		}()

		o.Stop()
	}, "Stop() should not panic")
}
