package cmd

import (
	"testing"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func init() {
	utils.SetStationName("tester")
}

func TestRunMQTTPub(t *testing.T) {
	// Mock the messanger.GetMQTT and its Publish method
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Define test cases
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{"ValidArgs", []string{"/test/topic", "test message"}, false},
		{"MissingArgs", []string{"/test/topic"}, true},
		{"NoArgs", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				runMsgPub(cmd, tt.args)
				return nil
			}()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if len(tt.args) > 1 {
				cli := mqtt.Client.(*messanger.MockClient)

				pub := cli.GetPublications()
				l := len(pub)
				p := pub[l-1]

				assert.Equal(t, "o/tester/"+tt.args[0], p.Topic)
				assert.Equal(t, tt.args[1], p.Payload)
			}
		})
	}
}
