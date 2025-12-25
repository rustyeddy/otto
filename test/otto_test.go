package system

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rustyeddy/otto/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	url    = "http://localhost:8011"
	broker = "otto"

	ocmd *exec.Cmd
)

func init() {
	os.Setenv("MQTT_BROKER", "otto")
	os.Setenv("MQTT_USER", "otto")
	os.Setenv("MQTT_PASS", "otto123")
}

func TestMain(m *testing.M) {
	m.Run()
}

func TestRunTests(t *testing.T) {
	t.Run("start", startOttO)
	t.Run("stop", stopOttO)
}

func startOttO(t *testing.T) {

	path, err := exec.LookPath("../otto")
	require.NoError(t, err, "expect to find the executable otto but did not: %s", path)

	ctx, _ := context.WithCancel(context.Background())
	ocmd = exec.CommandContext(ctx, "../otto", "serve")

	var stdout, stderr bytes.Buffer
	ocmd.Stdout = &stdout
	ocmd.Stderr = &stderr

	err = ocmd.Start()
	require.NoError(t, err, "expected to run otto but got an error")

	time.Sleep(1 * time.Second)

	// verify otto is running
	cli := client.NewClient(url)
	err = cli.Ping()
	assert.NoError(t, err, "expected no error but got one")

}

func stopOttO(t *testing.T) {
	ocmd.Cancel()

	stdout := ocmd.Stdout.(*bytes.Buffer)
	stderr := ocmd.Stderr.(*bytes.Buffer)

	os.WriteFile("stdout.log", stdout.Bytes(), 0644)
	os.WriteFile("stderr.log", stderr.Bytes(), 0644)
}
