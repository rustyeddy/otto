package main

import (
	"time"

	"github.com/rustyeddy/otto/data"
	"github.com/rustyeddy/otto/device/bme280"
	"github.com/rustyeddy/otto/messanger"
)

func main() {
	// create the topic the bme will publish to and the DataManager
	// will subscribe to
	topic := messanger.GetTopics().Data("bme280")

	// Set the BME i2c device and address Initialize the bme to use
	// the i2c bus
	bme := bme280.New("bme280", "/dev/i2c-1", 0x76)
	bme.Topic = topic
	err := bme.Init()
	if err != nil {
		panic(err)
	}

	// Before we start reading temp, etc. let's subscribe to
	// the messages we are going to publish.
	dm := data.GetDataManager()
	dm.Subscribe(topic, dm.Callback)

	// start reading in a loop and publish the results via MQTT
	done := make(chan any)
	go bme.TimerLoop(5*time.Second, done, bme.ReadPub)

	<-done
}
