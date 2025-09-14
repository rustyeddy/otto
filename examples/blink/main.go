/*
Blink sets up pin 6 for an LED and goes into an endless
toggle mode.
*/

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/rustyeddy/otto/device/drivers"
	"github.com/rustyeddy/otto/device/led"
	"github.com/rustyeddy/otto/messanger"
)

var (
	useMQTT bool
	pinid   int
	mock    string
	count   int
)

func init() {
	flag.BoolVar(&useMQTT, "mqtt", false, "Use mqtt or a timer")
	flag.IntVar(&pinid, "pin", 6, "The GPIO pin the LED is attached to")
	flag.StringVar(&mock, "mock", "", "mock gpio and/or mqtt")
}

func main() {
	flag.Parse()

	// Create the led, name it "led" and add a publish topic
	led, done := initLED("led", pinid)
	if useMQTT {
		domqtt(led)
	} else {
		dotimer(led, 1*time.Second, done)
		fmt.Println("LED will blink every second")
	}
	<-done
}

func initLED(name string, pin int) (*led.LED, chan any) {
	led := led.New(name, pin)
	led.SetTopic(messanger.GetTopics().Data(led.Device.Name()))
	done := make(chan any)
	return led, done
}

func domqtt(led *led.LED) {
	led.Subscribe(messanger.GetTopics().Control(led.Device.Name()), led.Callback)
}

func dotimer(led *led.LED, period time.Duration, done chan any) {
	count = 0
	led.TimerLoop(period, done, func() error {
		// led.Set(count % 2)
		count++
		return nil
	})
}

func domock() {
	switch mock {
	case "mqtt":
		messanger.NewMessanger("local", "")

	case "gpio":
		drivers.GetGPIO().Mock = true

	case "both":
		// messanger.SetMQTTClient(messanger.GetMockClient())
		messanger.NewMessanger("local", "")
		drivers.GetGPIO().Mock = true

	default:
		return
	}
}
