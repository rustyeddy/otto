package cmd

import (
	"fmt"
	"sync"
	"testing"

	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMsgSubCmdRegistration(t *testing.T) {
	// Test that msgSubCmd is registered as a subcommand of msgCmd
	found := false
	for _, cmd := range msgCmd.Commands() {
		if cmd.Use == "sub" {
			found = true
			break
		}
	}
	assert.True(t, found, "msgSubCmd should be registered as a subcommand of msgCmd")
}

func TestMsgSubCmdProperties(t *testing.T) {
	// Test command properties
	assert.Equal(t, "sub", msgSubCmd.Use)
	assert.Equal(t, "Subscribe to the msg topic", msgSubCmd.Short)
	assert.Equal(t, "Subscribe to msg tocpic", msgSubCmd.Long)
}

func TestRunMSGSubWithMockClient(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")
	require.NotNil(t, mqtt.Client)

	// Set client as connected
	mqtt.Connect()

	// Test successful subscription
	cmd := &cobra.Command{}
	args := []string{"sensors/temperature"}

	runMSGSub(cmd, args)

	// Verify subscription was recorded
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)

	sub := subscriptions["sensors/temperature"]
	assert.Equal(t, "sensors/temperature", sub.Topic)
	assert.NotNil(t, sub.Handler)
}

func TestRunMSGSubConnectionRequired(t *testing.T) {
	// Setup mock client that's initially disconnected
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Mock Connect to succeed and set connected state
	mqtt.Connect()

	// Test subscription with disconnected client
	cmd := &cobra.Command{}
	args := []string{"test/topic"}

	runMSGSub(cmd, args)

	// Verify Connect was called
	assert.True(t, mqtt.IsConnected(), "Connect should be called when client is not connected")

	// Verify subscription was made after connection
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)
	sub := subscriptions["test/topic"]
	assert.Equal(t, "test/topic", sub.Topic)
}

func TestRunMSGSubNilClient(t *testing.T) {
	// Setup MQTT with nil client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test subscription with nil client
	cmd := &cobra.Command{}
	args := []string{"test/topic"}

	runMSGSub(cmd, args)

	// Mock Connect to succeed and set mock client
	mqtt.Connect()

	// Verify Connect was called
	assert.True(t, mqtt.IsConnected(),
		"Connect should be called when client is nil")

	// Verify subscription was made after connection
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)
	assert.Equal(t, "test/topic", subscriptions["test/topic"].Topic)
}

func TestRunMSGSubAlreadyConnected(t *testing.T) {
	// Setup mock client that's already connected
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test subscription when already connected
	cmd := &cobra.Command{}
	args := []string{"test/topic"}

	runMSGSub(cmd, args)

	// Verify Connect was NOT called
	assert.False(t, mqtt.IsConnected(), "Connect should not be called when already connected")

	// Verify subscription was made
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)
	assert.Equal(t, "test/topic", subscriptions["test/topic"].Topic)
}

func TestRunMSGSubMultipleTopics(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test with multiple topics - should use first one
	cmd := &cobra.Command{}
	args := []string{"first/topic", "second/topic", "third/topic"}

	runMSGSub(cmd, args)

	// Verify only first topic was subscribed to
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)
	assert.Equal(t, "first/topic", subscriptions["first/topic"].Topic)
}

func TestRunMSGSubVariousTopicFormats(t *testing.T) {
	// Test various topic formats
	topicTests := []struct {
		name  string
		topic string
	}{
		{"Simple", "temperature"},
		{"Hierarchical", "sensors/building1/floor2/room3/temperature"},
		{"WithNumbers", "device123/sensor456"},
		{"WithDashes", "smart-home/living-room/temperature"},
		{"WithUnderscores", "iot_device_001/status"},
		{"MixedCase", "SmartHome/LivingRoom/Temperature"},
		{"Wildcards", "sensors/+/temperature"},
		{"MultiLevelWildcard", "sensors/#"},
	}

	for _, tt := range topicTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh mock client for each test
			mqtt := messanger.NewMessangerMQTT("test", "mock")

			runMSGSub(&cobra.Command{}, []string{tt.topic})

			cli := mqtt.Client.(*messanger.MockClient)
			subscriptions := cli.GetSubscriptions()
			require.Len(t, subscriptions, 1)
			assert.Equal(t, tt.topic, subscriptions[tt.topic].Topic)
		})
	}
}

func TestRunMSGSubMsgPrinterHandler(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test subscription
	cmd := &cobra.Command{}
	args := []string{"test/topic"}

	runMSGSub(cmd, args)

	// Verify handler is from MsgPrinter
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)

	sub := subscriptions["test/topic"]
	assert.NotNil(t, sub.Handler, "Handler should be set from MsgPrinter")

	// Test that handler can be called without panicking
	assert.NotPanics(t, func() {
		sub.Handler(cli, messanger.NewMockMessage("test/topic", []byte("test message")))
	}, "MsgPrinter handler should not panic")
}

func TestRunMSGSubMultipleSubscriptions(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Subscribe to multiple topics
	topics := []string{
		"sensors/temperature",
		"sensors/humidity",
		"alerts/fire",
		"status/online",
	}

	for _, topic := range topics {
		runMSGSub(&cobra.Command{}, []string{topic})
	}

	// Verify all subscriptions were made
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, len(topics))

	for _, topic := range topics {
		assert.Equal(t, topic, subscriptions[topic].Topic)
		assert.NotNil(t, subscriptions[topic].Handler)
	}
}

func TestRunMSGSubConcurrent(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test concurrent subscriptions
	const numGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			topic := fmt.Sprintf("test/routine%d", routineID)
			runMSGSub(&cobra.Command{}, []string{topic})
		}(i)
	}

	wg.Wait()

	// Verify all subscriptions were made
	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	assert.Len(t, subscriptions, numGoroutines)

	// Verify all topics are unique
	topicSet := make(map[string]bool)
	for _, sub := range subscriptions {
		topicSet[sub.Topic] = true
	}
	assert.Len(t, topicSet, numGoroutines, "All topics should be unique")
}

func TestMsgSubCmdIntegration(t *testing.T) {
	// Setup mock client
	mqtt := messanger.NewMessangerMQTT("test", "mock")

	// Test the command can be found and executed
	cmd, args, err := msgCmd.Find([]string{"sub", "integration/test"})
	require.NoError(t, err)
	assert.Equal(t, msgSubCmd, cmd)
	assert.Equal(t, []string{"integration/test"}, args)

	// Execute the command
	cmd.Run(cmd, args)

	cli := mqtt.Client.(*messanger.MockClient)
	subscriptions := cli.GetSubscriptions()
	require.Len(t, subscriptions, 1)

	sub := subscriptions["integration/test"]
	assert.Equal(t, "integration/test", sub.Topic)
	assert.NotNil(t, sub.Handler)
}
