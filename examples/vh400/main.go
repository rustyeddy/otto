package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rustyeddy/otto/device/vh400"
)

func main() {
	var err error
	pin := 0

	if len(os.Args) > 1 {
		pin, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("Bad argument %s - expected integers for adc\n", os.Args[1])
		}
	}

	readQ := make(<-chan float64)
	s := vh400.New("vh400", pin)
	readQ = s.AnalogPin.ReadContinuous()

	for {
		select {
		case val := <-readQ:
			fmt.Printf("adc: %d: %5.2f\n", pin, val)
		}
	}
}
