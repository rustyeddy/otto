package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/rustyeddy/otto/device/bme280"
	"github.com/rustyeddy/otto/device/button"
	"github.com/rustyeddy/otto/device/drivers"
	"github.com/rustyeddy/otto/device/led"
	"github.com/rustyeddy/otto/device/oled"
	"github.com/rustyeddy/otto/device/relay"
	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/station"
	"github.com/warthog618/go-gpiocdev"
)

type controller struct {
	*led.LED
	*relay.Relay
	*bme280.BME280
	*oled.OLED

	onButton  *button.Button
	offButton *button.Button

	*messanger.DataManager
	*station.StationManager
	*server.Server
}

func (c *controller) cleanup() {
	g := drivers.GetGPIO()
	g.Close()
}

func (c *controller) initSignals() {
	// todo make sure we capture signals
}

func (c *controller) initDataManager() {
	dm := messanger.GetDataManager()
	dm.Subscribe("ss/d/#", dm.Callback)

	c.Server.Register("/api/data", dm)
	c.DataManager = dm
}

func (c *controller) initStations() {
	sm := station.GetStationManager()
	sm.Start()
	c.StationManager = sm
	c.Server.Register("/api/stations", sm)
}

func (c *controller) initDevices(done chan any) error {
	c.initRelay(22)
	c.initLED(6)
	c.initButton("on", 23, gpiocdev.WithRisingEdge)
	c.initButton("off", 27, gpiocdev.WithFallingEdge)
	c.initBME280("/dev/i2c-1", 0x76, done)
	c.initOLED(done)

	messanger.GetMQTT().Subscribe(messanger.GetTopics().Control("button"), c.buttonCallback)
	return nil
}

func (c *controller) initRelay(idx int) {
	relay := relay.New("relay", idx)
	relay.Topic = messanger.GetTopics().Data("relay")
	relay.Subscribe(messanger.GetTopics().Control("relay"), relay.Callback)
	c.Relay = relay
}

func (c *controller) initLED(idx int) {
	led := led.New("led", idx)
	led.Topic = messanger.GetTopics().Data("led")
	led.Subscribe(messanger.GetTopics().Control("led"), led.Callback)
	c.LED = led
}

func (c *controller) initButton(name string, idx int, opts ...gpiocdev.LineReqOption) {
	but := button.New(name, idx, opts...)
	but.Topic = messanger.GetTopics().Control("button")
	go but.EventLoop(done, but.ReadPub)
	if name == "on" {
		c.onButton = but
	} else if name == "off" {
		c.offButton = but
	}
}

func (c *controller) initBME280(bus string, addr int, done chan any) (bme *bme280.BME280, err error) {
	bme = bme280.New("bme280", "/dev/i2c-1", 0x76)
	if bme == nil {
		return nil, fmt.Errorf("Failed initialize BME280 %s %d", "/dev/i2c-1", 0x76)
	}
	err = bme.Init()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	bme.Topic = messanger.GetTopics().Data("bme280")
	go bme.TimerLoop(10*time.Second, done, bme.ReadPub)
	c.BME280 = bme

	return bme, nil
}

func (c *controller) initOLED(done chan any) {

	display, err := oled.New("oled", 128, 64)
	if err != nil {
		log.Fatal(err)
	}

	dispvals := struct {
		temp     float64
		pressure float64
		humidity float64
		relay    string
	}{}

	refresh := func() {
		display.Clear()
		start := 25
		t := time.Now().Format(time.Kitchen)
		display.DrawString(10, 10, "OttO: "+t)
		display.DrawString(10, start, fmt.Sprintf("temp: %7.2f", dispvals.temp))
		display.DrawString(10, start+15, fmt.Sprintf("pres: %7.2f", dispvals.pressure))
		display.DrawString(10, start+30, fmt.Sprintf("humi: %7.2f", dispvals.humidity))
		display.Draw()
	}

	m := messanger.GetMQTT()
	display.Subscribe(messanger.GetTopics().Data("bme280"), func(msg *messanger.Msg) {
		mm, err := msg.Map()
		if err != nil {
			slog.Error("Failed top get map", "error", err)
		}

		var ex bool
		dispvals.temp, ex = mm["Temperature"].(float64)
		if !ex {
			slog.Error("failed to get temperature")
		}
		dispvals.pressure, ex = mm["Pressure"].(float64)
		if !ex {
			slog.Error("failed to get pressure")
		}
		dispvals.humidity, ex = mm["Humidity"].(float64)
		if !ex {
			slog.Error("failed to get Humidity")
		}
		refresh()
	})

	m.Subscribe(messanger.GetTopics().Data("relay"), func(msg *messanger.Msg) {
		dispvals.relay = msg.String()
	})
	c.OLED = display
}

func (c *controller) buttonCallback(msg *messanger.Msg) {
	cmd := msg.String()
	c.LED.PubData(cmd)
	c.Relay.PubData(cmd)
	return
}
