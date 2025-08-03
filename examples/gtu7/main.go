package main

import "github.com/rustyeddy/otto/device/gtu7"

func main() {
	g := gtu7.NewGTU7("/dev/ttyS0")
	gpsQ := g.StartReading()

	for gps := range gpsQ {
		g.PubData(gps)
	}
}
