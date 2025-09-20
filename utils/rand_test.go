package utils

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestNewRando(t *testing.T) {
	r := NewRando()

	if r == nil {
		t.Fatal("NewRando() returned nil")
	}

	if r.F != 0.0 {
		t.Errorf("Expected initial F value to be 0.0, got %f", r.F)
	}
}

func TestRandoFloat64(t *testing.T) {
	r := NewRando()

	// Test that Float64() returns values in valid range [0.0, 1.0)
	for i := 0; i < 100; i++ {
		v := r.Float64()

		if v < 0.0 || v >= 1.0 {
			t.Errorf("Float64() returned value %f outside expected range [0.0, 1.0)", v)
		}

		if math.IsNaN(v) {
			t.Error("Float64() returned NaN")
		}

		if math.IsInf(v, 0) {
			t.Error("Float64() returned infinity")
		}
	}
}

func TestRandoFloat64Uniqueness(t *testing.T) {
	r := NewRando()

	// Test that multiple calls return different values (with high probability)
	values := make(map[float64]bool)
	duplicates := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		v := r.Float64()
		if values[v] {
			duplicates++
		}
		values[v] = true
	}

	// Allow for some duplicates but expect most values to be unique
	// With Float64, duplicates should be extremely rare
	if duplicates > iterations/10 { // More than 10% duplicates is suspicious
		t.Errorf("Too many duplicate values: %d out of %d", duplicates, iterations)
	}
}

func TestRandoString(t *testing.T) {
	r := NewRando()

	for i := 0; i < 10; i++ {
		str := r.String()

		// Verify it returns a string interface
		if str == nil {
			t.Error("String() returned nil")
			continue
		}

		// Convert to actual string
		strValue, ok := str.(string)
		if !ok {
			t.Errorf("String() did not return a string type, got %T", str)
			continue
		}

		if strValue == "" {
			t.Error("String() returned empty string")
			continue
		}

		// Verify the string can be parsed as a float
		f, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			t.Errorf("String() returned unparseable float string: %s, error: %v", strValue, err)
			continue
		}

		// Verify the parsed float is in valid range
		if f < 0.0 || f >= 1.0 {
			t.Errorf("String() returned float %f outside expected range [0.0, 1.0)", f)
		}
	}
}

func TestRandoStringFormat(t *testing.T) {
	r := NewRando()
	str := r.String().(string)

	// Test that the string contains a decimal point (since it's formatted with %f)
	if !strings.Contains(str, ".") {
		t.Errorf("Expected string to contain decimal point, got: %s", str)
	}

	// Test that the string doesn't contain invalid characters
	for _, char := range str {
		if !((char >= '0' && char <= '9') || char == '.' || char == 'e' || char == 'E' || char == '+' || char == '-') {
			t.Errorf("String contains invalid character '%c' in: %s", char, str)
		}
	}
}

func TestRandoServeHTTP(t *testing.T) {
	r := NewRando()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call ServeHTTP
	r.ServeHTTP(w, req)

	resp := w.Result()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	resp.Body.Close()

	bodyStr := string(body)

	// Verify response is not empty
	if bodyStr == "" {
		t.Error("ServeHTTP returned empty response body")
	}

	// Verify response can be parsed as float
	f, err := strconv.ParseFloat(bodyStr, 64)
	if err != nil {
		t.Errorf("ServeHTTP returned unparseable float: %s, error: %v", bodyStr, err)
	}

	// Verify float is in valid range
	if f < 0.0 || f >= 1.0 {
		t.Errorf("ServeHTTP returned float %f outside expected range [0.0, 1.0)", f)
	}
}

func TestRandoServeHTTPMultipleRequests(t *testing.T) {
	r := NewRando()

	values := make([]float64, 0, 10)

	// Make multiple requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		body, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Fatalf("Failed to read response body for request %d: %v", i, err)
		}

		f, err := strconv.ParseFloat(string(body), 64)
		if err != nil {
			t.Errorf("Request %d returned unparseable float: %s", i, string(body))
			continue
		}

		values = append(values, f)
	}

	// Verify we got some variety (not all the same value)
	if len(values) > 1 {
		allSame := true
		first := values[0]
		for _, v := range values[1:] {
			if v != first {
				allSame = false
				break
			}
		}

		if allSame {
			t.Error("All HTTP responses returned the same value, randomness may be broken")
		}
	}
}

func TestRandoHTTPServer(t *testing.T) {
	// Test using httptest.NewServer (integration test)
	ts := httptest.NewServer(NewRando())
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Parse the response
	var f float64
	n, err := fmt.Sscanf(bodyStr, "%f", &f)
	if err != nil {
		t.Errorf("Failed to parse response as float: %v", err)
	}

	if n != 1 {
		t.Errorf("Expected to parse 1 float, got %d", n)
	}

	if f < 0 || f >= 1.0 {
		t.Errorf("Expected float in range [0.0, 1.0), got %f", f)
	}
}

func TestRandoHTTPDifferentMethods(t *testing.T) {
	r := NewRando()
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Method %s returned status %d, expected %d", method, w.Code, http.StatusOK)
			}

			body := w.Body.String()
			if body == "" {
				t.Errorf("Method %s returned empty body", method)
			}

			// Verify it's a valid float
			_, err := strconv.ParseFloat(body, 64)
			if err != nil {
				t.Errorf("Method %s returned invalid float: %s", method, body)
			}
		})
	}
}

// Benchmark tests
func BenchmarkRandoFloat64(b *testing.B) {
	r := NewRando()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Float64()
	}
}

func BenchmarkRandoString(b *testing.B) {
	r := NewRando()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.String()
	}
}

func BenchmarkRandoServeHTTP(b *testing.B) {
	r := NewRando()
	req := httptest.NewRequest("GET", "/", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// Test concurrent access (if needed in the future)
func TestRandoConcurrency(t *testing.T) {
	r := NewRando()
	done := make(chan bool, 10)

	// Run 10 goroutines concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine calls methods multiple times
			for j := 0; j < 100; j++ {
				_ = r.Float64()
				_ = r.String()

				// Also test HTTP handler
				req := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panic, the test passed
}
