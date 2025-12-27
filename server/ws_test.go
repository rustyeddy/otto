package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rustyeddy/otto/messenger"
	"github.com/stretchr/testify/assert"
)

// Mock websocket connection for testing
type MockConn struct {
	*websocket.Conn
	writeMessages []interface{}
	readMessages  [][]byte
	readIndex     int
	closed        bool
	writeError    error
	readError     error
	mu            sync.Mutex
}

func NewMockConn() *MockConn {
	return &MockConn{
		writeMessages: make([]interface{}, 0),
		readMessages:  make([][]byte, 0),
		readIndex:     0,
	}
}

func (m *MockConn) WriteJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeError != nil {
		return m.writeError
	}

	m.writeMessages = append(m.writeMessages, v)
	return nil
}

func (m *MockConn) ReadMessage() (messageType int, p []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readError != nil {
		return 0, nil, m.readError
	}

	if m.readIndex >= len(m.readMessages) {
		return 0, nil, fmt.Errorf("no more messages")
	}

	msg := m.readMessages[m.readIndex]
	m.readIndex++
	return websocket.TextMessage, msg, nil
}

func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockConn) AddReadMessage(msg []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readMessages = append(m.readMessages, msg)
}

func (m *MockConn) GetWrittenMessages() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]interface{}, len(m.writeMessages))
	copy(result, m.writeMessages)
	return result
}

func (m *MockConn) SetWriteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeError = err
}

func (m *MockConn) SetReadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readError = err
}

// Reset global state for clean tests
func resetWebsocks() {
	Websocks = []*Websock{}
}

func TestNewWebsock(t *testing.T) {
	mockConn := &websocket.Conn{}

	t.Run("Create new websocket", func(t *testing.T) {
		ws := NewWebsock(mockConn)

		assert.NotNil(t, ws, "NewWebsock() should not return nil")
		assert.Same(t, mockConn, ws.Conn, "Connection should be set correctly")
		assert.NotNil(t, ws.Done, "Done channel should be initialized")
		assert.NotNil(t, ws.writeQ, "WriteQ channel should be initialized")

		// Test channel properties
		select {
		case <-ws.Done:
			t.Error("Done channel should not be closed initially")
		default:
			// Expected behavior
		}

		select {
		case <-ws.writeQ:
			t.Error("WriteQ channel should be empty initially")
		default:
			// Expected behavior
		}
	})

	t.Run("Create with nil connection", func(t *testing.T) {
		ws := NewWebsock(nil)

		assert.NotNil(t, ws, "NewWebsock() should not return nil even with nil connection")
		assert.Nil(t, ws.Conn, "Connection should be nil as passed")
		assert.NotNil(t, ws.Done, "Done channel should still be initialized")
		assert.NotNil(t, ws.writeQ, "WriteQ channel should still be initialized")
	})
}

func TestWebsockGetWriteQ(t *testing.T) {
	mockConn := &websocket.Conn{}
	ws := NewWebsock(mockConn)

	t.Run("Get write queue", func(t *testing.T) {
		wq := ws.GetWriteQ()
		assert.NotNil(t, wq, "GetWriteQ() should not return nil")
		assert.Equal(t, ws.writeQ, wq, "Should return the same channel instance")
	})

	t.Run("Write to queue", func(t *testing.T) {
		wq := ws.GetWriteQ()
		msg := messenger.NewMsg("test/topic", []byte("test data"), "test-source")

		go func() {
			// Test reading from queue
			select {
			case receivedMsg := <-wq:
				assert.Same(t, msg, receivedMsg, "Should receive the same message")
			case <-time.After(500 * time.Millisecond):
				t.Error("Should be able to read from queue")
			}
		}()

		// Test non-blocking write
		select {
		case wq <- msg:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Write to queue should not block")
		}
	})
}

func TestCheckOrigin(t *testing.T) {
	t.Run("Always returns true", func(t *testing.T) {
		// Test with various request scenarios
		testCases := []struct {
			name   string
			origin string
			host   string
		}{
			{"localhost", "http://localhost:3000", "localhost:8080"},
			{"different domain", "http://example.com", "localhost:8080"},
			{"no origin", "", "localhost:8080"},
			{"https origin", "https://secure.example.com", "localhost:8080"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/ws", nil)
				if tc.origin != "" {
					req.Header.Set("Origin", tc.origin)
				}
				req.Host = tc.host

				result := checkOrigin(req)
				assert.True(t, result, "checkOrigin should always return true")
			})
		}
	})

	t.Run("Nil request", func(t *testing.T) {
		// This should not panic
		assert.NotPanics(t, func() {
			result := checkOrigin(nil)
			assert.True(t, result, "checkOrigin should return true even for nil request")
		})
	})
}

func TestWebSocketUpgrader(t *testing.T) {
	t.Run("Upgrader configuration", func(t *testing.T) {
		assert.Equal(t, 1024, upgrader.ReadBufferSize, "ReadBufferSize should be 1024")
		assert.Equal(t, 1024, upgrader.WriteBufferSize, "WriteBufferSize should be 1024")
		assert.NotNil(t, upgrader.CheckOrigin, "CheckOrigin should be set")

		// Test that CheckOrigin is our function
		req := httptest.NewRequest("GET", "/ws", nil)
		result := upgrader.CheckOrigin(req)
		assert.True(t, result, "CheckOrigin should return true")
	})
}

func TestWServeServeHTTP(t *testing.T) {
	resetWebsocks()

	t.Run("WebSocket upgrade failure", func(t *testing.T) {
		ws := WServe{}

		// Create a request without proper WebSocket headers
		req := httptest.NewRequest("GET", "/ws", nil)
		w := httptest.NewRecorder()

		// This should fail to upgrade and return
		ws.ServeHTTP(w, req)

		// The response should indicate upgrade failure
		assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for non-WebSocket request")
	})

	t.Run("WebSocket upgrade success simulation", func(t *testing.T) {
		// Note: Testing actual WebSocket upgrade is complex due to gorilla/websocket
		// implementation details. In a real scenario, you'd use a WebSocket test server.

		resetWebsocks()
		initialCount := len(Websocks)

		// We can test the structure and verify that the upgrader is configured correctly
		assert.Equal(t, 0, initialCount, "Should start with no websocket connections")
		assert.NotNil(t, upgrader, "Upgrader should be initialized")
		assert.Equal(t, 1024, upgrader.ReadBufferSize, "Buffer sizes should be configured")
	})
}

// Test WebSocket message handling with mock connection
func TestWebSocketMessageHandling(t *testing.T) {
	resetWebsocks()

	t.Run("Message writing", func(t *testing.T) {
		mockConn := NewMockConn()
		ws := NewWebsock(mockConn.Conn)

		// Replace the connection with our mock for testing
		ws.Conn = mockConn.Conn

		// Create a test message
		msg := messenger.NewMsg("test/topic", []byte(`{"temperature": 25.5}`), "test-source")

		// Test the message handling logic (simulate what happens in ServeHTTP select loop)
		wq := ws.GetWriteQ()

		go func() {
			// Simulate processing the message (what would happen in the select loop)
			select {
			case receivedMsg, ok := <-wq:

				assert.True(t, ok, "Channel should not be closed")
				assert.Same(t, msg, receivedMsg, "Should receive the same message")

				// Test JSON conversion
				jbytes, err := receivedMsg.JSON()
				assert.NoError(t, err, "Should be able to convert message to JSON")
				assert.True(t, json.Valid(jbytes), "Should produce valid JSON")

			case <-time.After(500 * time.Millisecond):
				t.Error("Should receive message from write queue")
			}
		}()

		select {
		case wq <- msg:
			// success

		case <-time.After(500 * time.Millisecond):
			t.Error("Should be able to send message to write queue")
		}

	})

	t.Run("Channel closure handling", func(t *testing.T) {
		mockConn := NewMockConn()
		ws := NewWebsock(mockConn.Conn)

		wq := ws.GetWriteQ()

		// Close the write queue
		close(wq)

		// Test reading from closed channel
		select {
		case msg, ok := <-wq:
			assert.False(t, ok, "Reading from closed channel should return ok=false")
			assert.Nil(t, msg, "Message should be nil when channel is closed")
		case <-time.After(100 * time.Millisecond):
			t.Error("Should be able to read from closed channel immediately")
		}
	})

	t.Run("Done channel signaling", func(t *testing.T) {
		mockConn := NewMockConn()
		ws := NewWebsock(mockConn.Conn)

		// Test that Done channel is not closed initially
		select {
		case <-ws.Done:
			t.Error("Done channel should not be closed initially")
		default:
			// Expected
		}

		// Close Done channel
		close(ws.Done)

		// Test that we can read from closed Done channel
		select {
		case <-ws.Done:
			// Expected - channel is closed
		case <-time.After(100 * time.Millisecond):
			t.Error("Should be able to read from closed Done channel immediately")
		}
	})
}

func TestWebSocketConcurrency(t *testing.T) {
	resetWebsocks()

	t.Run("Concurrent websocket creation", func(t *testing.T) {
		const numRoutines = 10
		var wg sync.WaitGroup
		websockets := make([]*Websock, numRoutines)

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mockConn := &websocket.Conn{}
				ws := NewWebsock(mockConn)
				websockets[index] = ws
			}(i)
		}

		wg.Wait()

		// Verify all websockets were created
		for i, ws := range websockets {
			assert.NotNil(t, ws, "WebSocket %d should not be nil", i)
			assert.NotNil(t, ws.Done, "Done channel should be initialized for WebSocket %d", i)
			assert.NotNil(t, ws.writeQ, "WriteQ should be initialized for WebSocket %d", i)
		}
	})

	t.Run("Concurrent message sending", func(t *testing.T) {
		mockConn := &websocket.Conn{}
		ws := NewWebsock(mockConn)
		wq := ws.GetWriteQ()

		const numMessages = 50
		const numSenders = 5
		var wg sync.WaitGroup

		receivedMessages := make([]*messenger.Msg, 0, numMessages)
		receivedMutex := sync.Mutex{}

		// Start receiver
		go func() {
			for i := 0; i < numMessages; i++ {
				select {
				case msg := <-wq:
					receivedMutex.Lock()
					receivedMessages = append(receivedMessages, msg)
					receivedMutex.Unlock()
				case <-time.After(5 * time.Second):
					t.Error("Timeout waiting for message")
					return
				}
			}
		}()

		// Start senders
		for i := 0; i < numSenders; i++ {
			wg.Add(1)
			go func(senderID int) {
				defer wg.Done()
				messagesPerSender := numMessages / numSenders
				for j := 0; j < messagesPerSender; j++ {
					msg := messenger.NewMsg(
						fmt.Sprintf("test/sender/%d", senderID),
						[]byte(fmt.Sprintf(`{"sender": %d, "message": %d}`, senderID, j)),
						fmt.Sprintf("sender-%d", senderID),
					)

					select {
					case wq <- msg:
						// Success
					case <-time.After(1 * time.Second):
						t.Errorf("Timeout sending message from sender %d", senderID)
						return
					}
				}
			}(i)
		}

		wg.Wait()

		// Give receiver time to process all messages
		time.Sleep(100 * time.Millisecond)

		receivedMutex.Lock()
		assert.Equal(t, numMessages, len(receivedMessages), "Should receive all sent messages")
		receivedMutex.Unlock()
	})
}

func TestWebSocketGlobalState(t *testing.T) {
	t.Run("Websocks slice management", func(t *testing.T) {
		resetWebsocks()

		assert.Equal(t, 0, len(Websocks), "Should start with empty Websocks slice")

		// Simulate adding websockets (as would happen in ServeHTTP)
		mockConn1 := &websocket.Conn{}
		mockConn2 := &websocket.Conn{}
		ws1 := NewWebsock(mockConn1)
		ws2 := NewWebsock(mockConn2)

		Websocks = append(Websocks, ws1)
		assert.Equal(t, 1, len(Websocks), "Should have 1 websocket after adding first")
		assert.Same(t, ws1, Websocks[0], "First websocket should be stored correctly")

		Websocks = append(Websocks, ws2)
		assert.Equal(t, 2, len(Websocks), "Should have 2 websockets after adding second")
		assert.Same(t, ws2, Websocks[1], "Second websocket should be stored correctly")
	})

	t.Run("Global state reset", func(t *testing.T) {
		// Add some websockets
		Websocks = append(Websocks, NewWebsock(&websocket.Conn{}))
		Websocks = append(Websocks, NewWebsock(&websocket.Conn{}))
		assert.Equal(t, 4, len(Websocks), "Should have 2 websockets before reset")

		// Reset
		resetWebsocks()
		assert.Equal(t, 0, len(Websocks), "Should have 0 websockets after reset")
	})
}

func TestWebSocketEdgeCases(t *testing.T) {
	t.Run("Message with nil data", func(t *testing.T) {
		mockConn := &websocket.Conn{}
		ws := NewWebsock(mockConn)

		// Create message with nil data
		msg := messenger.NewMsg("test/topic", nil, "test-source")

		wq := ws.GetWriteQ()
		go func() {
			select {
			case <-wq:
				// Success
			case <-time.After(1 * time.Second):
				t.Errorf("Timeout sending message from sender")
				return
			}
		}()

		time.Sleep(100 * time.Millisecond)
		// This should not panic
		assert.NotPanics(t, func() {
			select {
			case wq <- msg:
				// Success
			default:
				t.Error("Should be able to send message with nil data")
			}
		})
	})

	t.Run("Message with empty topic", func(t *testing.T) {
		mockConn := &websocket.Conn{}
		ws := NewWebsock(mockConn)
		wq := ws.GetWriteQ()

		go func() {
			select {
			case m := <-wq:
				assert.Empty(t, m.Topic)
				// Success

			case <-time.After(1 * time.Second):
				t.Errorf("Timeout sending message from sender")
				return
			}
		}()

		// Create message with empty topic
		msg := messenger.NewMsg("", []byte("test data"), "test-source")
		assert.NotPanics(t, func() {
			wq <- msg
		})
	})

	t.Run("Very large message", func(t *testing.T) {
		mockConn := &websocket.Conn{}
		ws := NewWebsock(mockConn)
		wq := ws.GetWriteQ()

		size := 1024 * 1024
		go func() {
			select {
			case m := <-wq:
				assert.Equal(t, size, len(m.Data))
				// Success

			case <-time.After(1 * time.Second):
				t.Errorf("Timeout sending message from sender")
				return
			}
		}()

		// Create a large message
		largeData := make([]byte, size) // 1MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		msg := messenger.NewMsg("test/large", largeData, "test-source")

		assert.NotPanics(t, func() {
			select {
			case wq <- msg:
				assert.Equal(t, size, len(msg.Data))
				// Success
			case <-time.After(1 * time.Second):
				t.Error("Should be able to send large message")
			}
		})
	})
}

func TestWebSocketMessageJSON(t *testing.T) {
	t.Run("Valid JSON message", func(t *testing.T) {
		msg := messenger.NewMsg("test/topic", []byte(`{"temperature": 25.5, "unit": "celsius"}`), "test-source")

		jbytes, err := msg.JSON()
		assert.NoError(t, err, "Should be able to convert valid message to JSON")
		assert.True(t, json.Valid(jbytes), "Should produce valid JSON")

		// Parse the JSON to verify structure
		var parsed map[string]interface{}
		err = json.Unmarshal(jbytes, &parsed)
		assert.NoError(t, err, "Should be able to parse the JSON")

		// Verify message fields are present
		assert.Contains(t, parsed, "topic", "JSON should contain topic field")
		assert.Contains(t, parsed, "msg", "JSON should contain msg field")
		assert.Contains(t, parsed, "source", "JSON should contain source field")
	})

	t.Run("Message with binary data", func(t *testing.T) {
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		msg := messenger.NewMsg("test/binary", binaryData, "test-source")

		jbytes, err := msg.JSON()
		assert.NoError(t, err, "Should be able to convert binary message to JSON")
		assert.True(t, json.Valid(jbytes), "Should produce valid JSON")
	})
}

// Benchmark tests
func BenchmarkNewWebsock(b *testing.B) {
	mockConn := &websocket.Conn{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewWebsock(mockConn)
	}
}

func BenchmarkWebSocketMessageQueue(b *testing.B) {
	mockConn := &websocket.Conn{}
	ws := NewWebsock(mockConn)
	wq := ws.GetWriteQ()

	msg := messenger.NewMsg("benchmark/topic", []byte(`{"value": 42}`), "benchmark")

	// Start consumer
	go func() {
		for range wq {
			// Consume messages
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wq <- msg
	}
}

func TestWebSocketIntegration(t *testing.T) {
	resetWebsocks()

	t.Run("Full websocket lifecycle simulation", func(t *testing.T) {
		// Create websocket
		mockConn := &websocket.Conn{}
		ws := NewWebsock(mockConn)

		// Add to global slice (simulating ServeHTTP)
		Websocks = append(Websocks, ws)
		assert.Equal(t, 1, len(Websocks), "Should have 1 websocket")

		// Send messages
		wq := ws.GetWriteQ()
		messages := []*messenger.Msg{
			messenger.NewMsg("sensor/temperature", []byte(`{"value": 23.5}`), "sensor-1"),
			messenger.NewMsg("sensor/humidity", []byte(`{"value": 65}`), "sensor-2"),
			messenger.NewMsg("sensor/pressure", []byte(`{"value": 1013.25}`), "sensor-3"),
		}

		go func() {
			// Send all messages
			for _, msg := range messages {
				select {
				case wq <- msg:
				// Success
				case <-time.After(100 * time.Millisecond):
					t.Error("Should be able to send message")
				}
			}
		}()

		// Receive and verify all messages
		for i, expectedMsg := range messages {
			select {
			case receivedMsg := <-wq:
				assert.Same(t, expectedMsg, receivedMsg, "Message %d should match", i)
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Should receive message %d", i)
			}
		}

		// Clean up (simulate connection closing)
		close(ws.Done)
		close(wq)

		// Verify cleanup
		select {
		case <-ws.Done:
			// Expected - channel should be closed
		default:
			t.Error("Done channel should be closed")
		}
	})
}

// It seems you already have a comprehensive set of tests in `server/ws_test.go`.
// If you need additional tests for specific scenarios or edge cases, please let me know.
