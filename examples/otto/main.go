package main

import (
	"embed"
	"flag"

	"github.com/rustyeddy/otto/device/drivers"
	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/utils"
	"github.com/spf13/cobra-cli/cmd"
)

var (
	done     chan any
	loglevel string
	mock     bool
	mockMQTT bool
	mockGPIO bool

	cli bool
)

func init() {
	flag.BoolVar(&cli, "cli", false, "Run the otto interactive cli")
	flag.BoolVar(&mockMQTT, "mock", false, "mock the hardware")
	flag.BoolVar(&mockMQTT, "mock-mqtt", false, "mock the hardware")
	flag.BoolVar(&mockGPIO, "mock-gpio", false, "mock the hardware")
	flag.StringVar(&loglevel, "log", "warn", "Logging level debug, info, warn, error, fatal")
}

func main() {
	flag.Parse()

	done = make(chan any)
	initLogging()
	initMock()

	// TODO capture signals
	c := &controller{}
	c.initApp()
	c.initSignals()
	c.initStations()
	c.initDataManager()
	c.initDevices(done)

	if cli {
		cmd.Execute()
	}

	<-done
	c.cleanup()
}

func initLogging() {

	level := loglevel
	logfile := "otto.log"
	utils.InitLogger(level, logfile)
}

func initMock() {
	if mock {
		mockMQTT = true
		mockGPIO = true
	}
	if mockMQTT {
		messanger.SetMQTTClient(messanger.GetMockClient())
	}
	if mockGPIO {
		drivers.GetGPIO().Mock = true
	}
}

//go:embed app
var content embed.FS

func (c *controller) initApp() {
	done := make(chan any)

	s := server.GetServer()

	// The following line is commented out because
	s.EmbedTempl("/emb", content, nil)
	s.Appdir("/", "app")
	s.Start(done)
	c.Server = s
	<-done
}
