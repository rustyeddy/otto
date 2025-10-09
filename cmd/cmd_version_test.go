package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmdRegistration(t *testing.T) {
	// Test that versionCmd is registered with rootCmd
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("versionCmd should be registered with rootCmd")
	}
}

func TestVersionCmdProperties(t *testing.T) {
	// Test command properties
	if versionCmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short != "Print the version number of otto" {
		t.Errorf("expected Short to be 'Print the version number of otto', got '%s'", versionCmd.Short)
	}

	if versionCmd.Long != "All software has versions. This is OttO's" {
		t.Errorf("expected Long to be 'All software has versions. This is OttO's', got '%s'", versionCmd.Long)
	}
}

func TestVersionValue(t *testing.T) {
	// Test that version variable has expected value
	if version != "0.1.0" {
		t.Errorf("expected version to be '0.1.0', got '%s'", version)
	}
}

func TestVersionCmdRun(t *testing.T) {
	// Capture output
	output := new(bytes.Buffer)
	originalOutput := cmdOutput
	cmdOutput = output
	defer func() { cmdOutput = originalOutput }()

	// Execute version command
	cmd := &cobra.Command{}
	args := []string{}
	versionCmd.Run(cmd, args)

	// Verify output
	expectedOutput := version + "\n"
	if output.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, output.String())
	}
}

func TestVersionCmdWithArgs(t *testing.T) {
	// Test that args are ignored
	var output bytes.Buffer
	originalOutput := cmdOutput
	cmdOutput = &output
	defer func() { cmdOutput = originalOutput }()

	// Execute with various args
	testArgs := [][]string{
		{"arg1"},
		{"arg1", "arg2"},
		{"--flag"},
		{"multiple", "arguments", "here"},
	}

	for _, args := range testArgs {
		output.Reset()
		versionCmd.Run(&cobra.Command{}, args)

		expectedOutput := version + "\n"
		if output.String() != expectedOutput {
			t.Errorf("expected output '%s' with args %v, got '%s'", expectedOutput, args, output.String())
		}
	}
}

func TestVersionCmdOutputWriter(t *testing.T) {
	// Test with different output writers
	writers := []io.Writer{
		&bytes.Buffer{},
		os.Stdout,
		io.Discard,
	}

	originalOutput := cmdOutput
	defer func() { cmdOutput = originalOutput }()

	for i, writer := range writers {
		t.Run(fmt.Sprintf("Writer%d", i), func(t *testing.T) {
			cmdOutput = writer

			// Should not panic with different writers
			assert.NotPanics(t, func() {
				versionCmd.Run(&cobra.Command{}, []string{})
			})
		})
	}
}

func TestVersionCmdIntegration(t *testing.T) {
	// Test finding and executing the version command through rootCmd
	cmd, args, err := rootCmd.Find([]string{"version"})
	if err != nil {
		t.Fatalf("expected to find version command, got error: %v", err)
	}

	if cmd != versionCmd {
		t.Error("expected to find versionCmd")
	}

	if len(args) != 0 {
		t.Errorf("expected no remaining args, got %v", args)
	}

	// Capture output
	var output bytes.Buffer
	originalOutput := cmdOutput
	cmdOutput = &output
	defer func() { cmdOutput = originalOutput }()

	// Execute the found command
	cmd.Run(cmd, args)

	// Verify output
	expectedOutput := version + "\n"
	if output.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, output.String())
	}
}

func TestVersionModification(t *testing.T) {
	// Test behavior when version variable is modified
	originalVersion := version
	defer func() { version = originalVersion }()

	testVersions := []string{
		"1.0.0",
		"2.5.3-beta",
		"dev-build",
		"",
	}

	for _, testVersion := range testVersions {
		t.Run(fmt.Sprintf("Version_%s", testVersion), func(t *testing.T) {
			version = testVersion

			var output bytes.Buffer
			originalOutput := cmdOutput
			cmdOutput = &output
			defer func() { cmdOutput = originalOutput }()

			versionCmd.Run(&cobra.Command{}, []string{})

			expectedOutput := testVersion + "\n"
			if output.String() != expectedOutput {
				t.Errorf("expected output '%s', got '%s'", expectedOutput, output.String())
			}
		})
	}
}
