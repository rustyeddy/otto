package otto

import (
	"testing"
)

func TestOttOInit(t *testing.T) {
	o := &OttO{Name: "TestOttO"}

	o.Init()

	if o.done == nil {
		t.Error("Expected done channel to be initialized")
	}

	if o.Messanger == nil {
		t.Error("Expected Messanger to be initialized")
	}

	if o.StationManager == nil {
		t.Error("Expected StationManager to be initialized")
	}

	if o.Station == nil {
		t.Error("Expected Station to be initialized")
	}

	if o.Server == nil {
		t.Error("Expected Server to be initialized")
	}

	if o.brokerShutdown == nil {
		t.Error("Expected brokerShutdown function to be initialized")
	}
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
