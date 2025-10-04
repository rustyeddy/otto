package data

// func TestFileStorage(t *testing.T) {
// 	// Create a temporary directory for testing
// 	tempDir := t.TempDir()

// 	t.Run("Create File Storage", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		assert.Equal(t, tempDir, storage.Filename)
// 	})

// 	t.Run("Save and Load Sensor Data", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		sensorData := &SensorData{
// 			ID:        "test-sensor-001",
// 			Type:      "temperature",
// 			Value:     23.5,
// 			Unit:      "celsius",
// 			Timestamp: time.Now().Truncate(time.Second),
// 			Quality:   "good",
// 		}
// 		ts := NewTimeseries("sensors")
// 		ts.Add(sensorData)

// 		// Save the data
// 		err := storage.Save("sensors", ts)
// 		require.NoError(t, err)

// 		// Load the data back
// 		ts, err = storage.Load("sensors")
// 		require.NoError(t, err)

// 		loaded := ts.Datas[0].(SensorData)
// 		require.NotNil(t, loaded)

// 		assert.Equal(t, sensorData.ID, loaded.ID)
// 		assert.Equal(t, sensorData.Type, loaded.Type)
// 		assert.Equal(t, sensorData.Value, loaded.Value)
// 		assert.Equal(t, sensorData.Unit, loaded.Unit)
// 		assert.Equal(t, sensorData.Quality, loaded.Quality)
// 		assert.True(t, sensorData.Timestamp.Equal(loaded.Timestamp))
// 	})

// 	t.Run("Save and Load Multiple Readings", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		readings := []*WeatherReading{
// 			{
// 				ID:          "multi-test-001",
// 				Timestamp:   time.Now().Add(-2 * time.Hour).Truncate(time.Second),
// 				Temperature: 22.0,
// 				Humidity:    60.0,
// 			},
// 			{
// 				ID:          "multi-test-001",
// 				Timestamp:   time.Now().Add(-1 * time.Hour).Truncate(time.Second),
// 				Temperature: 24.0,
// 				Humidity:    58.0,
// 			},
// 		}

// 		// Save multiple readings
// 		ts := NewTimeseries("weather")
// 		for _, reading := range readings {
// 			ts.Add(reading)
// 			require.NoError(t, err)
// 		}
// 		err := storage.Save("weather", tsg)

// 		// Load all readings for the station
// 		loaded, err := storage.Load("weather")
// 		require.NoError(t, err)
// 		assert.Len(t, loaded, 2)

// 		// Verify the loaded data
// 		assert.Equal(t, 22.0, loaded[0].Temperature)
// 		assert.Equal(t, 24.0, loaded[1].Temperature)
// 	})

// 	t.Run("Load Non-existent Data", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		// Try to load non-existent sensor data
// 		data, err := storage.LoadSensorData("non-existent")
// 		assert.Error(t, err)
// 		assert.Nil(t, data)

// 		// Try to load readings for non-existent station
// 		readings, err := storage.LoadWeatherReadings("non-existent-station")
// 		assert.Error(t, err)
// 		assert.Nil(t, readings)
// 	})

// 	t.Run("List Stored Data", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		// Save some test data
// 		sensors := []string{"sensor-001", "sensor-002", "sensor-003"}
// 		for _, id := range sensors {
// 			data := &SensorData{
// 				ID:        id,
// 				Type:      "temperature",
// 				Value:     20.0,
// 				Timestamp: time.Now(),
// 			}
// 			err := storage.SaveSensorData(data)
// 			require.NoError(t, err)
// 		}

// 		// List all sensor IDs
// 		ids, err := storage.ListSensorIDs()
// 		require.NoError(t, err)
// 		assert.Len(t, ids, 3)

// 		for _, expectedID := range sensors {
// 			assert.Contains(t, ids, expectedID)
// 		}
// 	})

// 	t.Run("Delete Data", func(t *testing.T) {
// 		storage := NewFileStore(tempDir)

// 		// Save some data first
// 		data := &SensorData{
// 			ID:        "delete-test-001",
// 			Type:      "humidity",
// 			Value:     65.0,
// 			Timestamp: time.Now(),
// 		}

// 		err := storage.SaveSensorData(data)
// 		require.NoError(t, err)

// 		// Verify it exists
// 		loaded, err := storage.LoadSensorData("delete-test-001")
// 		require.NoError(t, err)
// 		assert.NotNil(t, loaded)

// 		// Delete it
// 		err = storage.DeleteSensorData("delete-test-001")
// 		require.NoError(t, err)

// 		// Verify it's gone
// 		deleted, err := storage.LoadSensorData("delete-test-001")
// 		assert.Error(t, err)
// 		assert.Nil(t, deleted)
// 	})
// }

// func TestMemoryStorage(t *testing.T) {
// 	t.Run("Create Memory Storage", func(t *testing.T) {
// 		storage := NewMemoryStorage()

// 		assert.NotNil(t, storage.sensors)
// 		assert.NotNil(t, storage.weather)
// 		assert.NotNil(t, storage.mu)
// 		assert.Empty(t, storage.sensors)
// 		assert.Empty(t, storage.weather)
// 	})

// 	t.Run("Store and Retrieve Sensor Data", func(t *testing.T) {
// 		storage := NewMemoryStorage()

// 		data := &SensorData{
// 			ID:        "memory-test-001",
// 			Type:      "pressure",
// 			Value:     1013.25,
// 			Unit:      "hPa",
// 			Timestamp: time.Now(),
// 		}

// 		// Store data
// 		err := storage.SaveSensorData(data)
// 		require.NoError(t, err)

// 		// Retrieve data
// 		retrieved, err := storage.LoadSensorData("memory-test-001")
// 		require.NoError(t, err)
// 		assert.Equal(t, data.ID, retrieved.ID)
// 		assert.Equal(t, data.Value, retrieved.Value)
// 	})

// 	t.Run("Store and Retrieve Weather Data", func(t *testing.T) {
// 		storage := NewMemoryStorage()

// 		reading := &WeatherReading{
// 			StationID:   "memory-weather-001",
// 			Timestamp:   time.Now(),
// 			Temperature: 25.5,
// 			Humidity:    70.0,
// 		}

// 		// Store reading
// 		err := storage.SaveWeatherReading(reading)
// 		require.NoError(t, err)

// 		// Retrieve readings
// 		readings, err := storage.LoadWeatherReadings("memory-weather-001")
// 		require.NoError(t, err)
// 		assert.Len(t, readings, 1)
// 		assert.Equal(t, 25.5, readings[0].Temperature)
// 	})

// 	t.Run("Memory Storage Thread Safety", func(t *testing.T) {
// 		storage := NewMemoryStorage()

// 		// Test concurrent writes and reads
// 		done := make(chan bool, 4)

// 		// Writer goroutine 1
// 		go func() {
// 			defer func() { done <- true }()
// 			for i := 0; i < 10; i++ {
// 				data := &SensorData{
// 					ID:    fmt.Sprintf("concurrent-1-%d", i),
// 					Value: float64(i),
// 				}
// 				storage.SaveSensorData(data)
// 			}
// 		}()

// 		// Writer goroutine 2
// 		go func() {
// 			defer func() { done <- true }()
// 			for i := 0; i < 10; i++ {
// 				reading := &WeatherReading{
// 					StationID:   fmt.Sprintf("concurrent-station-%d", i),
// 					Temperature: float64(i * 2),
// 				}
// 				storage.SaveWeatherReading(reading)
// 			}
// 		}()

// 		// Reader goroutine 1
// 		go func() {
// 			defer func() { done <- true }()
// 			for i := 0; i < 10; i++ {
// 				storage.LoadSensorData(fmt.Sprintf("concurrent-1-%d", i))
// 				time.Sleep(1 * time.Millisecond)
// 			}
// 		}()

// 		// Reader goroutine 2
// 		go func() {
// 			defer func() { done <- true }()
// 			for i := 0; i < 10; i++ {
// 				storage.LoadWeatherReadings(fmt.Sprintf("concurrent-station-%d", i))
// 				time.Sleep(1 * time.Millisecond)
// 			}
// 		}()

// 		// Wait for all goroutines
// 		for i := 0; i < 4; i++ {
// 			select {
// 			case <-done:
// 			case <-time.After(5 * time.Second):
// 				t.Fatal("Goroutine did not complete in time")
// 			}
// 		}

// 		// Verify some data was stored
// 		ids, err := storage.ListSensorIDs()
// 		require.NoError(t, err)
// 		assert.True(t, len(ids) > 0, "Should have stored some sensor data")
// 	})

// 	t.Run("Clear Memory Storage", func(t *testing.T) {
// 		storage := NewMemoryStorage()

// 		// Add some data
// 		storage.SaveSensorData(&SensorData{ID: "clear-test-1"})
// 		storage.SaveWeatherReading(&WeatherReading{StationID: "clear-station-1"})

// 		// Verify data exists
// 		ids, _ := storage.ListSensorIDs()
// 		assert.True(t, len(ids) > 0)

// 		// Clear storage
// 		storage.Clear()

// 		// Verify it's empty
// 		ids, _ = storage.ListSensorIDs()
// 		assert.Empty(t, ids)

// 		readings, _ := storage.LoadWeatherReadings("clear-station-1")
// 		assert.Empty(t, readings)
// 	})
// }

// func TestDataCompression(t *testing.T) {
// 	t.Run("Compress and Decompress JSON", func(t *testing.T) {
// 		original := &WeatherReading{
// 			StationID:     "compress-test-001",
// 			Timestamp:     time.Now(),
// 			Temperature:   23.5,
// 			Humidity:      65.2,
// 			Pressure:      1013.25,
// 			WindSpeed:     12.3,
// 			WindDirection: 180.0,
// 			Rainfall:      2.5,
// 		}

// 		// Convert to JSON
// 		jsonData, err := json.Marshal(original)
// 		require.NoError(t, err)

// 		// Compress
// 		compressed, err := CompressData(jsonData)
// 		require.NoError(t, err)
// 		assert.True(t, len(compressed) < len(jsonData), "Compressed data should be smaller")

// 		// Decompress
// 		decompressed, err := DecompressData(compressed)
// 		require.NoError(t, err)

// 		// Verify it matches original JSON
// 		assert.Equal(t, jsonData, decompressed)

// 		// Verify we can unmarshal back to the original struct
// 		var restored WeatherReading
// 		err = json.Unmarshal(decompressed, &restored)
// 		require.NoError(t, err)

// 		assert.Equal(t, original.StationID, restored.StationID)
// 		assert.Equal(t, original.Temperature, restored.Temperature)
// 		assert.Equal(t, original.Humidity, restored.Humidity)
// 	})

// 	t.Run("Compression Ratio", func(t *testing.T) {
// 		// Create a large dataset to test compression effectiveness
// 		readings := make([]WeatherReading, 100)
// 		for i := 0; i < 100; i++ {
// 			readings[i] = WeatherReading{
// 				StationID:     fmt.Sprintf("station-%03d", i%10), // Repeated station IDs
// 				Timestamp:     time.Now().Add(time.Duration(i) * time.Minute),
// 				Temperature:   20.0 + float64(i%20),
// 				Humidity:      50.0 + float64(i%50),
// 				Pressure:      1000.0 + float64(i%30),
// 				WindSpeed:     float64(i % 25),
// 				WindDirection: float64((i * 15) % 360),
// 				Rainfall:      float64(i%5) * 0.5,
// 			}
// 		}

// 		jsonData, err := json.Marshal(readings)
// 		require.NoError(t, err)

// 		compressed, err := CompressData(jsonData)
// 		require.NoError(t, err)

// 		compressionRatio := float64(len(compressed)) / float64(len(jsonData))
// 		t.Logf("Original size: %d bytes", len(jsonData))
// 		t.Logf("Compressed size: %d bytes", len(compressed))
// 		t.Logf("Compression ratio: %.2f", compressionRatio)

// 		assert.True(t, compressionRatio < 0.8, "Should achieve at least 20% compression")
// 	})
// }
