package main

import (
	"fmt"
	"time"

	"github.com/rustyeddy/otto/device/drivers"
	"periph.io/x/conn/v3/physic"
)

const (
	adcRange              = 26400 // ADS1115 gives you 32768 steps with 4.096v, with 3.3v which is the 80% of 4.096v we get 26400 steps
	inputVoltageInChannel = 3300 * physic.MilliVolt
)

func main() {
	ads := drivers.NewADS1115("ADS1115", "/dev/i2c-1", 0x48)
	ads.Init()

	var err error
	var pins [4]drivers.AnalogPin
	var chans4 [4]<-chan float64
	for i := 0; i < 4; i++ {
		pname := fmt.Sprintf("pin%d", i)
		pins[i], err = ads.Pin(pname, i, nil)
		if err != nil {
			fmt.Printf("Error creating pin: %d\n", i)
		}
		chans4[i] = pins[i].ReadContinuous()
	}

	for j := 0; j < 4; j++ {
		for i := 0; i < 4; i++ {
			val, err := pins[i].Read()
			if err != nil {
				fmt.Printf("failed to read pin[%d] = %s\n", i, err)
				continue
			}
			fmt.Printf("reading[%d]: %f\n", i, val)
		}
		time.Sleep(2 * time.Second)
	}

	var val float64
	for {
		select {
		case val = <-chans4[0]:
		case val = <-chans4[1]:
		case val = <-chans4[2]:
		case val = <-chans4[3]:
		}
	}
	fmt.Printf("VAL: %5.2f\n", val)
	ads.Close()

}
