/*
Blink sets up pin 6 for an LED and goes into an endless
toggle mode.
*/

package main

import (
	"log/slog"

	"encoding/json"

	"github.com/rustyeddy/otto/device/drivers"
)

var gpioStr = `
{
    "chipname":"gpiochip4",
    "pins": {
        "6": {
            "name": "led",
            "offset": 6,
            "value": 0,
            "mode": 0
        }
    }
}
`

func main() {
	var g drivers.GPIO
	if err := json.Unmarshal([]byte(gpioStr), &g); err != nil {
		slog.Error(err.Error())
		return
	}

	defer func() {
		g.Close()
	}()

	// TODO
	// led := g.pins[6]
	// for {
	// 	led.Toggle()
	// 	time.Sleep(1 * time.Second)
	// }
}
