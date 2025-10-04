package data

import (
	"encoding/json"
	"testing"
	"time"
)

type WeatherReading struct {
	ID            string
	Timestamp     time.Time
	Temperature   float64
	Humidity      float64
	Pressure      float64
	WindSpeed     float64
	WindDirection float64
}

func BenchmarkSensorDataJSON(b *testing.B) {
	data := &SensorData{
		ID:        "benchmark-sensor-001",
		Type:      "temperature",
		Value:     23.5,
		Unit:      "celsius",
		Timestamp: time.Now(),
		Quality:   "excellent",
	}

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	jsonData, _ := json.Marshal(data)
	b.Run("Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result SensorData
			err := json.Unmarshal(jsonData, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkWeatherDataProcessing(b *testing.B) {
	// Create test data
	readings := make([]*WeatherReading, 1000)
	for i := 0; i < 1000; i++ {
		readings[i] = &WeatherReading{
			Timestamp:     time.Now().Add(time.Duration(i) * time.Minute),
			Temperature:   20.0 + float64(i%20),
			Humidity:      50.0 + float64(i%50),
			Pressure:      1000.0 + float64(i%30),
			WindSpeed:     float64(i % 25),
			WindDirection: float64((i * 15) % 360),
		}
	}

	b.Run("AddReadings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			history := NewTimeseries("bench-station")
			for _, reading := range readings {
				history.Add(reading)
			}
		}
	})

	history := NewTimeseries("bench-station")
	for _, reading := range readings {
		history.Add(reading)
	}

	b.Run("TimeRangeQuery", func(b *testing.B) {
		start := time.Now().Add(-500 * time.Minute)
		end := time.Now()
		for i := 0; i < b.N; i++ {
			history.GetReadingsInRange(start, end)
		}
	})
}
