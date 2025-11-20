package utils

import "runtime"

type Stats struct {
	Goroutines int
	CPUs       int
	runtime.MemStats
	GoVersion string
}

func GetStats() *Stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s := &Stats{
		Goroutines: runtime.NumGoroutine(),
		CPUs:       runtime.NumCPU(),
		MemStats:   m,
		GoVersion:  runtime.Version(),
	}

	return s
}
