package utils

import "os"

var (
	stationName string
)

func SetStationName(name string) {
	stationName = name
}

func StationName() string {
	if stationName == "" {
		var err error
		stationName, err = os.Hostname()
		if err != nil {
			panic(err)
		}
	}
	return stationName
}
