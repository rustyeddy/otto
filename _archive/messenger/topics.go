package messenger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/rustyeddy/otto/utils"
)

// Topics manages topic formatting and usage tracking for the Otto messaging system.
// It provides helper methods for creating properly formatted topics and tracks
// how many times each topic has been used.
//
// Otto uses a standardized topic format: "ss/[c|d]/station/sensor"
// Where:
//   - "o" is the namespace prefix (Smart Station)
//   - "c" indicates control topics (commands to devices)
//   - "d" indicates data topics (sensor readings, telemetry)
//   - station is the station/device identifier
//   - sensor is the sensor/actuator/command name
//
// Example topics:
//   - "ss/c/station1/led" - Control topic to turn LED on/off
//   - "ss/d/station1/temp" - Data topic for temperature readings
type Topics struct {
	Prefix string         // Prefix for all topics
	Format string         // Format string for topic generation (e.g., "ss/%s/%s/%s")
	Map    map[string]int // Map of topic to usage count

	mu *sync.RWMutex `json:"-"`
}

var (
	topics *Topics
)

func init() {
	var mu sync.RWMutex
	topics = &Topics{
		Prefix: "o/",
		Format: "%s/%s/%s",
		Map:    make(map[string]int),
		mu:     &mu,
	}
	topics.Format = topics.Prefix + topics.Format
}

// ValidateTopic checks if a topic string follows Otto's topic format conventions.
// A valid Otto topic must have:
//   - At least 4 segments separated by '/'
//   - First segment must be "ss" (namespace)
//   - Second segment must be "c" (control) or "d" (data)
//   - Third segment (station ID) must not be empty
//   - Fourth segment (sensor/command) must not be empty
//
// Parameters:
//   - topic: The topic string to validate (e.g., "ss/c/station1/temp")
//
// Returns true if the topic is valid, false otherwise.
//
// Example:
//
//	valid := ValidateTopic("ss/c/station1/temp")  // Returns true
//	valid := ValidateTopic("invalid/topic")       // Returns false
func ValidateTopic(topic string) bool {
	path := strings.Split(topic, "/")
	if len(path) < 4 {
		return false
	}

	if path[0] != "o" {
		return false
	}

	if path[1] != "c" && path[1] != "d" {
		return false
	}

	if path[2] == "" || path[3] == "" {
		return false
	}

	// here we have to accept the station id and topic it advertises
	// because we can't know what the station IDs are.
	return true
}

// GetTopics returns the singleton Topics instance for the application.
// This provides access to topic formatting and usage tracking.
//
// Returns a pointer to the Topics instance.
//
// Example:
//
//	topics := GetTopics()
//	controlTopic := topics.Control("led")
func GetTopics() *Topics {
	return topics
}

// Control generates a control topic for the current station.
// Control topics are used to send commands to devices (e.g., turn on LED, set speed).
// The format is: "ss/c/{station}/{topic}"
//
// This method also increments the usage counter for the generated topic.
//
// Parameters:
//   - topic: The command or actuator name (e.g., "led", "motor", "relay")
//
// Returns the fully formatted control topic string.
//
// Example:
//
//	topics := GetTopics()
//	ledTopic := topics.Control("led")  // Returns "ss/c/mystation/led"
func (t *Topics) Control(topic string) string {
	top := fmt.Sprintf(t.Format, "c", utils.StationName(), topic)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Map[top]++
	return top
}

func ControlTopic(topic string) string {
	return topics.Control(topic)
}

// Data generates a data topic for the current station.
// Data topics are used to publish sensor readings and telemetry.
// The format is: "ss/d/{station}/{topic}"
//
// This method also increments the usage counter for the generated topic.
//
// Parameters:
//   - topic: The sensor or data stream name (e.g., "temp", "humidity", "motion")
//
// Returns the fully formatted data topic string.
//
// Example:
//
//	topics := GetTopics()
//	tempTopic := topics.Data("temp")  // Returns "ss/d/mystation/temp"
func (t *Topics) Data(topic string) string {
	top := fmt.Sprintf(t.Format, "d", utils.StationName(), topic)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Map[top]++
	return top
}

func DataTopic(topic string) string {
	return topics.Data(topic)
}

func Topic(topic string) string {
	top := topics.Prefix + topic
	topics.mu.Lock()
	defer topics.mu.Unlock()
	topics.Map[top]++
	return top
}

// ServeHTTP implements http.Handler to provide a REST API endpoint for
// inspecting topic usage. Returns JSON containing the topic format and
// a map of all topics used by this station with their usage counts.
//
// Response format:
//
//	{
//	  "TopicFmt": "ss/%s/%s/%s",
//	  "Topicmap": {
//	    "ss/c/station1/led": 5,
//	    "ss/d/station1/temp": 120
//	  }
//	}
//
// This is useful for monitoring which topics are being used and how frequently.
func (t *Topics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	t.mu.Lock()
	defer t.mu.Unlock()
	err := json.NewEncoder(w).Encode(t)
	if err != nil {
		slog.Error("Error wrote data", "error", err)
	}
}
