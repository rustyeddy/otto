package otto

import (
	"testing"

	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
)

func TestMqttCmd(t *testing.T) {
	cmd := mqttCmd
	if cmd == nil {
		t.Fatal("mqttCmd is nil")
	}
	if cmd.Use != "mqtt" {
		t.Errorf("expected Use to be 'mqtt', got %s", cmd.Use)
	}
	if cmd.Short != "Configure and interact with MQTT broker" {
		t.Errorf("expected Short to be 'Configure and interact with MQTT broker', got %s", cmd.Short)
	}
	if cmd.Run == nil {
		t.Errorf("expected Run to be defined, got nil")
	}
}

func TestMqttRun(t *testing.T) {
	mqttConfig.Broker = "localhost"
	cmd := &cobra.Command{}
	args := []string{}

	mqttRun(cmd, args)

	m := messanger.GetMQTT()
	// url := "tcp://localhost:1883"
	url := "localhost"
	if m.Broker != url {
		t.Errorf("expected Broker (%s) got (%s)", url, m.Broker)
	}
	if !m.IsConnected() {
		t.Error("expected MQTT to be connected")
	}
}
