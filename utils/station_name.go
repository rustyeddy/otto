package utils

var (
	stationName string
)

func SetStationName(name string) {
	stationName = name
}

func StationName() string {
	return stationName
}
