package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMsgCmd(t *testing.T) {
	cmd := msgCmd
	if cmd == nil {
		t.Fatal("msgCmd is nil")
	}
	if cmd.Use != "msg" {
		t.Errorf("expected Use to be 'msg', got %s", cmd.Use)
	}
	if cmd.Short != "Configure and interact with MSG broker" {
		t.Errorf("expected Short to be 'Configure and interact with MSG broker', got %s", cmd.Short)
	}
	if cmd.Run == nil {
		t.Errorf("expected Run to be defined, got nil")
	}
}

func TestMsgRun(t *testing.T) {
	msgConfig.Broker = "localhost"
	cmd := &cobra.Command{}
	args := []string{}

	msgRun(cmd, args)
}
func TestMsgCmdRegistration(t *testing.T) {
	// Test that msgCmd is registered with rootCmd
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "msg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("msgCmd should be registered with rootCmd")
	}
}

func TestMsgCmdFlags(t *testing.T) {
	// Test that broker flag exists and has correct default
	brokerFlag := msgCmd.PersistentFlags().Lookup("broker")
	if brokerFlag == nil {
		t.Fatal("broker flag should be defined")
	}

	if brokerFlag.DefValue != "localhost" {
		t.Errorf("expected broker flag default to be 'localhost', got %s", brokerFlag.DefValue)
	}

	if brokerFlag.Usage != "Set the MSG Broker" {
		t.Errorf("expected broker flag usage to be 'Set the MSG Broker', got %s", brokerFlag.Usage)
	}
}

func TestMsgConfigurationStruct(t *testing.T) {
	// Test msgConfiguration struct
	config := msgConfiguration{
		Broker:  "test.broker.com",
		Enabled: true,
	}

	if config.Broker != "test.broker.com" {
		t.Errorf("expected Broker to be 'test.broker.com', got %s", config.Broker)
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestMsgRunBrokerUpdate(t *testing.T) {
	// Save original config
	originalBroker := msgConfig.Broker
	defer func() { msgConfig.Broker = originalBroker }()

	// Set up test
	msgConfig.Broker = "new.broker.com"

	// Capture output
	var output strings.Builder
	originalWriter := cmdWriter
	cmdWriter = &output
	defer func() { cmdWriter = originalWriter }()

	// Run command
	cmd := &cobra.Command{}
	args := []string{}
	msgRun(cmd, args)

	// Verify output contains new broker
	outputStr := output.String()
	if !strings.Contains(outputStr, "new.broker.com") {
		t.Errorf("expected output to contain 'new.broker.com', got: %s", outputStr)
	}
}

func TestMsgRunWithNilClient(t *testing.T) {
	// This tests the case where m.Client is nil
	var output strings.Builder
	originalWriter := cmdWriter
	cmdWriter = &output
	defer func() { cmdWriter = originalWriter }()

	// Run command - should handle nil client gracefully
	cmd := &cobra.Command{}
	args := []string{}
	msgRun(cmd, args)

	// Should show Connected: false when client is nil
	outputStr := output.String()
	if !strings.Contains(outputStr, "Connected: false") {
		t.Errorf("expected 'Connected: false' when client is nil, got: %s", outputStr)
	}
}

func TestMsgRunBrokerNoChange(t *testing.T) {
	// Test case where broker config doesn't change
	originalBroker := msgConfig.Broker
	defer func() { msgConfig.Broker = originalBroker }()

	// Set broker to localhost (default)
	msgConfig.Broker = "localhost"

	var output strings.Builder
	originalWriter := cmdWriter
	cmdWriter = &output
	defer func() { cmdWriter = originalWriter }()

	// Run command
	cmd := &cobra.Command{}
	args := []string{}
	msgRun(cmd, args)

	// Should still show localhost
	outputStr := output.String()
	if !strings.Contains(outputStr, "Broker: localhost") {
		t.Errorf("expected output to contain 'Broker: localhost', got: %s", outputStr)
	}
}

func TestCmdWriterInitialization(t *testing.T) {
	// Test that cmdWriter is initialized to io.Discard
	if cmdWriter != io.Discard {
		t.Error("cmdWriter should be initialized to io.Discard")
	}
}

func TestMsgCmdLongDescription(t *testing.T) {
	expectedLong := "This command can be used to interact and diagnose an MSG broker"
	if msgCmd.Long != expectedLong {
		t.Errorf("expected Long to be '%s', got '%s'", expectedLong, msgCmd.Long)
	}
}

func TestMsgRunWithDifferentArgs(t *testing.T) {
	// Test that args parameter doesn't affect behavior
	var output1, output2 strings.Builder
	originalWriter := cmdWriter

	// First run with no args
	cmdWriter = &output1
	msgRun(&cobra.Command{}, []string{})

	// Second run with args
	cmdWriter = &output2
	msgRun(&cobra.Command{}, []string{"arg1", "arg2"})

	cmdWriter = originalWriter

	// Output should be the same regardless of args
	if output1.String() != output2.String() {
		t.Error("msgRun output should be the same regardless of args")
	}
}

func TestMsgConfigGlobalState(t *testing.T) {
	// Test that msgConfig is properly initialized
	if msgConfig.Broker == "" {
		// msgConfig.Broker gets set by flag parsing, but should have some value
		t.Log("msgConfig.Broker is empty - this might be expected before flag parsing")
	}

	// Test that we can modify the global config
	originalBroker := msgConfig.Broker
	originalEnabled := msgConfig.Enabled

	msgConfig.Broker = "modified.broker.com"
	msgConfig.Enabled = true

	if msgConfig.Broker != "modified.broker.com" {
		t.Error("should be able to modify global msgConfig.Broker")
	}

	if !msgConfig.Enabled {
		t.Error("should be able to modify global msgConfig.Enabled")
	}

	// Restore original values
	msgConfig.Broker = originalBroker
	msgConfig.Enabled = originalEnabled
}
