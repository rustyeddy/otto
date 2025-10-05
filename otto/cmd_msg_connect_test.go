package otto

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRunMQTTConnect(t *testing.T) {
	// Mock the command and arguments
	cmd := &cobra.Command{}
	args := []string{}

	// Call the function
	runMQTTConnect(cmd, args)

	// Since runMQTTConnect calls messanger.GetMQTT(), you would typically
	// mock messanger.GetMQTT() and verify it was called. However, without
	// more details about the messanger package, this test is limited to
	// ensuring no errors occur during execution.
	t.Log("runMQTTConnect executed successfully")
}
