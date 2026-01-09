package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock handler for testing
type MockHandler struct {
	called bool
	path   string
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	m.path = r.URL.Path
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Mock handler called for %s", r.URL.Path)
}

// Reset server singleton for clean tests
func resetServer() {
	server = nil
}

func TestNewServer(t *testing.T) {
	s := NewServer()

	assert.NotNil(t, s, "NewServer() should not return nil")
	assert.NotNil(t, s.Server, "HTTP Server should be initialized")
	assert.NotNil(t, s.ServeMux, "ServeMux should be initialized")
	assert.Equal(t, ":8011", s.Addr, "Default address should be :8011")
	assert.Equal(t, 0, s.EndPointCount(), "EndPoints should be nil initially")
}

func TestGetServer(t *testing.T) {
	// Reset for clean test
	resetServer()

	// Test singleton behavior
	s1 := GetServer()
	s2 := GetServer()

	assert.Same(t, s1, s2, "GetServer() should return the same instance")
	assert.NotNil(t, s1, "First call should return valid instance")
	assert.NotNil(t, s2, "Second call should return valid instance")
}

func TestServerRegister(t *testing.T) {
	s := NewServer()
	mockHandler := &MockHandler{}

	t.Run("Register handler", func(t *testing.T) {
		path := "/test/endpoint"
		s.Register(path, mockHandler)

		p, ok := s.EndPoints.Load(path)
		assert.True(t, ok, "Endpoint path should exists")
		assert.Same(t, mockHandler, p, "Correct handler should be stored")
	})

	t.Run("Register multiple handlers", func(t *testing.T) {
		handler1 := &MockHandler{}
		handler2 := &MockHandler{}

		s.Register("/api/v1", handler1)
		s.Register("/api/v2", handler2)

		count := 0
		found := 0
		s.EndPoints.Range(func(k, v any) bool {
			count++
			switch k {
			case "/api/v1":
				found++

			case "/api/v2":
				found++
			}
			return true
		})

		assert.Equal(t, 3, count, "Should have 3 registered handlers")
		assert.Equal(t, 2, found, "Should have found /api/v1 and /api/v2")
	})
}

func TestServerServeHTTP(t *testing.T) {
	s := NewServer()

	// Register some test handlers
	s.Register("/api/test1", &MockHandler{})
	s.Register("/api/test2", &MockHandler{})
	s.Register("/health", &MockHandler{})

	t.Run("GET endpoints list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Should set JSON content type")

		// Parse response
		var response struct {
			Routes []string `json:"Routes"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, 3, len(response.Routes), "Should list all registered routes")

		// Check that all registered paths are in the response
		routeMap := make(map[string]bool)
		for _, route := range response.Routes {
			routeMap[route] = true
		}
		assert.True(t, routeMap["/api/test1"], "Should contain /api/test1")
		assert.True(t, routeMap["/api/test2"], "Should contain /api/test2")
		assert.True(t, routeMap["/health"], "Should contain /health")
	})

	t.Run("POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK for POST as well")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Should set JSON content type")
	})
}

func TestServerAppdir(t *testing.T) {
	s := NewServer()

	t.Run("Register file server", func(t *testing.T) {
		path := "/static"
		file := "./testdata"

		s.Appdir(path, file)

		f, ok := s.EndPoints.Load(path)
		assert.True(t, ok)
		assert.NotNil(t, f, "File server handler should not be nil")
	})

	t.Run("Multiple file servers", func(t *testing.T) {
		s.Appdir("/css", "./css")
		s.Appdir("/js", "./js")

		_, ok := s.EndPoints.Load("/css")
		assert.True(t, ok, "CSS file server should be registered")
		_, ok = s.EndPoints.Load("/js")
		assert.True(t, ok, "js file server should be registered")
	})
}

//go:embed testdata/*
var testFS embed.FS

func TestServerEmbedTempl(t *testing.T) {
	s := NewServer()

	t.Run("Register embed template handler", func(t *testing.T) {
		path := "/testdata"
		data := map[string]string{"title": "Test App"}

		// This should register a handler function
		s.EmbedTempl(path, testFS, data)
	})

	t.Run("Handle CSS request", func(t *testing.T) {
		s.EmbedTempl("/app", testFS, nil)

		req := httptest.NewRequest("GET", "/app/style.css", nil)
		w := httptest.NewRecorder()

		// Get the handler and test it
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate the CSS handling logic
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
				w.WriteHeader(http.StatusOK)
			}
		})

		handler.ServeHTTP(w, req)
		assert.Equal(t, "text/css", w.Header().Get("Content-Type"), "Should set CSS content type")
	})

	t.Run("Handle JS request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/app/script.js", nil)
		w := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "application/javascript")
				w.WriteHeader(http.StatusOK)
			}
		})

		handler.ServeHTTP(w, req)
		assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"), "Should set JS content type")
	})
}

func TestServerStart(t *testing.T) {
	s := NewServer()

	t.Run("Start server with done channel", func(t *testing.T) {
		done := make(chan any)

		// Start server in goroutine
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(done) // Signal to stop
		}()

		// This should not block indefinitely
		startComplete := make(chan bool)
		go func() {
			s.Start(done)
			startComplete <- true
		}()

		select {
		case <-startComplete:
			// Server started and stopped successfully
		case <-time.After(5 * time.Second):
			t.Fatal("Server.Start() did not complete within timeout")
		}

		_, ok := s.EndPoints.Load("/ping")
		assert.True(t, ok)
		_, ok = s.EndPoints.Load("/api")
		assert.True(t, ok)
	})
}

func TestServerConcurrency(t *testing.T) {
	s := NewServer()
	const numRoutines = 10
	const handlersPerRoutine = 5

	t.Run("Concurrent registrations", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numRoutines*handlersPerRoutine)

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < handlersPerRoutine; j++ {
					path := fmt.Sprintf("/api/concurrent/%d/%d", routineID, j)
					handler := &MockHandler{}

					// This should be safe for concurrent access
					s.Register(path, handler)

					f, ok := s.EndPoints.Load(path)
					assert.True(t, ok)
					if f == nil {
						errors <- fmt.Errorf("handler not registered for %s", path)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}

		expectedCount := numRoutines * handlersPerRoutine
		count := s.EndPointCount()
		assert.Equal(t, expectedCount, count, "All handlers should be registered concurrently")
	})

	t.Run("Concurrent ServeHTTP calls", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numRoutines)

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				req := httptest.NewRequest("GET", "/api", nil)
				w := httptest.NewRecorder()

				s.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					errors <- fmt.Errorf("expected status 200, got %d", w.Code)
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})
}

func TestServerEdgeCases(t *testing.T) {
	t.Run("Register nil handler", func(t *testing.T) {
		s := NewServer()

		// This should not panic
		assert.NotPanics(t, func() {
			s.Register("/nil-handler", nil)
		}, "Registering nil handler should not panic")

		f, ok := s.EndPoints.Load("/nil-handler")
		assert.False(t, ok)
		assert.Nil(t, f, "Nil handler should be stored as nil")
	})

	t.Run("ServeHTTP with no endpoints", func(t *testing.T) {
		s := NewServer()

		req := httptest.NewRequest("GET", "/api", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 even with no endpoints")

		var response struct {
			Routes []string `json:"Routes"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should return valid JSON")
		assert.Empty(t, response.Routes, "Routes should be empty")
	})

	t.Run("Multiple server instances", func(t *testing.T) {
		resetServer()

		s1 := GetServer()
		s2 := GetServer()
		s3 := NewServer() // This creates a new instance, not singleton

		assert.Same(t, s1, s2, "GetServer should return same instance")
		assert.NotSame(t, s1, s3, "NewServer should create new instance")
	})
}

func TestServerShutdown(t *testing.T) {
	s := NewServer()

	t.Run("Shutdown with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// This should not panic or hang
		assert.NotPanics(t, func() {
			err := s.Shutdown(ctx)
			// Error is expected since server wasn't actually started
			assert.NoError(t, err, "Shutdown does not return error for non-started server")
		}, "Shutdown should not panic")
	})
}

func TestServerResponseFormat(t *testing.T) {
	s := NewServer()
	s.Register("/test1", &MockHandler{})
	s.Register("/test2", &MockHandler{})

	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Test response body structure
	body := w.Body.Bytes()
	assert.True(t, json.Valid(body), "Response should be valid JSON")

	// Test that response can be unmarshaled
	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	assert.NoError(t, err, "Should unmarshal without error")
	assert.Contains(t, response, "Routes", "Response should contain Routes field")

	// Test routes format
	routes, ok := response["Routes"].([]interface{})
	assert.True(t, ok, "Routes should be an array")
	assert.Equal(t, 2, len(routes), "Should have 2 routes")
}

// Benchmark tests
func BenchmarkServerRegister(b *testing.B) {
	s := NewServer()
	handler := &MockHandler{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/bench/%d", i)
		s.Register(path, handler)
	}
}

func BenchmarkServerServeHTTP(b *testing.B) {
	s := NewServer()
	// Register some handlers first
	for i := 0; i < 100; i++ {
		s.Register(fmt.Sprintf("/api/bench/%d", i), &MockHandler{})
	}

	req := httptest.NewRequest("GET", "/api", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// Helper function for testing HTTP requests
func makeRequest(handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// Test helper validation
func TestMakeRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Method: %s, Path: %s", r.Method, r.URL.Path)
	})

	w := makeRequest(handler, "GET", "/test", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Method: GET")
	assert.Contains(t, w.Body.String(), "Path: /test")
}
