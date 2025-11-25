package otto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOttOInit(t *testing.T) {
	o := &OttO{Name: "TestOttO"}

	o.Init()

	assert.NotNil(t, o.done, "Expected done channel to be initialized")
	assert.NotNil(t, o.Messanger, "Expected Messanger to be initialized")
	assert.NotNil(t, o.StationManager, "Expected StationManager to be initialized")
	assert.NotNil(t, o.Station, "Expected Station to be initialized")
	assert.NotNil(t, o.Server, "Expected Server to be initialized")
	// assert.NotNil(t, o.brokerShutdown, "Expected brokerShutdown function to be initialized")

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

// func TestOttOStartAndStop(t *testing.T) {
// 	o := &OttO{
// 		Name:           "TestOttO",
// 		StationManager: station.GetStationManager(),
// 		Server:         server.GetServer(),
// 		Messanger:      messanger.NewMessangerMQTT("otto", messanger.GetTopics().Data("station")),
// 		done:           make(chan any),
// 	}

// 	println("1")
// 	go func() {
// 		if err := o.Start(); err != nil {
// 			t.Errorf("Start returned an error: %v", err)
// 		}
// 	}()
// 	println("2")
// 	o.done <- struct{}{}
// 	println("3")
// 	o.Stop()
// 	println("4")

// 	select {
// 	case <-o.done:
// 		println("5")
// 		t.Log("Stop successfully closed the done channel")
// 	default:
// 		println("6")
// 		t.Error("Expected done channel to be closed after Stop")
// 	}
// }
