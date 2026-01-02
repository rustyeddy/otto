package system

import (
	"bytes"
	"context"
	"flag"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rustyeddy/otto/client"
	"github.com/rustyeddy/otto/messenger"
	"github.com/rustyeddy/otto/station"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Config struct {
	URL          string
	Broker       string
	StationCount int
	OttOStart    bool
}

var (
	config Config
	ocmd   *exec.Cmd
)

type systemTest struct {
	*messenger.Messenger
	*messenger.MsgSaver
	*station.StationManager
}

func init() {
	os.Setenv("MQTT_BROKER", "localhost")
	os.Setenv("MQTT_USER", "otto")
	os.Setenv("MQTT_PASS", "otto123")

	flag.BoolVar(&config.OttOStart, "ottostart", false, "Have the test start auto otherwise it looks for an existing otto")
	flag.StringVar(&config.URL, "url", "http://localhost:8011", "URL for OttO")
	flag.StringVar(&config.Broker, "broker", "localhost", "Host name for MQTT broker")
	flag.IntVar(&config.StationCount, "station-count", 10, "Station Count")
}

func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
}

func TestRunTests(t *testing.T) {
	st := &systemTest{}

	t.Run("start", st.startOttO)
	t.Run("messenger", st.startMessenger)
	t.Run("stations", st.testStations)
	// t.Run("stations")

	time.Sleep(10 * time.Minute)
	t.Run("stop", st.stopOttO)
}

func (ts *systemTest) startOttO(t *testing.T) {
	if config.OttOStart {
		path, err := exec.LookPath("../otto")
		require.NoError(t, err, "expect to find the executable otto but did not: %s", path)

		ctx, _ := context.WithCancel(context.Background())
		ocmd = exec.CommandContext(ctx, "../otto", "serve")

		var stdout, stderr bytes.Buffer
		ocmd.Stdout = &stdout
		ocmd.Stderr = &stderr

		err = ocmd.Start()
		require.NoError(t, err, "expected to run otto but got an error")

		time.Sleep(500 * time.Millisecond)
	}

	// verify otto is running
	cli := client.NewClient(config.URL)
	err := cli.Ping()
	assert.NoError(t, err, "expected no error but got one")
}

func (ts *systemTest) stopOttO(t *testing.T) {
	ocmd.Cancel()

	stdout := ocmd.Stdout.(*bytes.Buffer)
	stderr := ocmd.Stderr.(*bytes.Buffer)

	os.WriteFile("stdout.log", stdout.Bytes(), 0644)
	os.WriteFile("stderr.log", stderr.Bytes(), 0644)

	if ts.MsgSaver != nil {
		ts.MsgSaver.StopSaving()
	}
}

func (ts *systemTest) startMessenger(t *testing.T) {
	msgr := messenger.NewMessenger(config.Broker)
	msgr.Connect()
	msgr.Pub("o/hello", "world")

	ts.MsgSaver = messenger.GetMsgSaver()
	ts.MsgSaver.StartSaving()
}
