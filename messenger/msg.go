package messenger

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rustyeddy/otto/utils"
)

// Msg represents a message in the Otto messaging system. It contains the message
// payload along with routing information and metadata for tracking and debugging.
//
// Each message has a globally unique ID and timestamp for tracing message flow
// through the system. The topic is parsed into a path array for efficient routing.
//
// Fields:
//   - ID: Unique monotonically increasing identifier for this message
//   - Topic: Full topic path (e.g., "ss/c/station/temp")
//   - Path: Topic split into segments for routing (e.g., ["ss", "c", "station", "temp"])
//   - Args: Optional arguments extracted from the topic path
//   - Data: The message payload as a byte array
//   - Source: Identifier of the component that created this message
//   - Valid: Whether the topic follows Otto's topic format conventions
//   - Timestamp: When the message was created
//
// Messages are immutable once created and should be created using NewMsg().
type Msg struct {
	ID     int64    `json:"id"`
	Topic  string   `json:"topic"`
	Path   []string `json:"path"`
	Args   []string `json:"args"`
	Data   []byte   `json:"msg"`
	Source string   `json:"source"`
	Valid  bool     `json:"valid"`

	Timestamp time.Duration `json:"timestamp"`
}

var (
	// msgid is the global message ID counter. It's incremented for each new message.
	msgid    int64 = 0
	msgSaver *MsgSaver
)

// getMsgID returns a globally unique message ID by incrementing a counter.
// Each message created in the system gets a unique sequential ID.
//
// Note: This is not thread-safe. In a concurrent environment, proper
// synchronization should be added to prevent ID collisions.
//
// Returns the next message ID.
func getMsgID() int64 {
	msgid++
	return msgid
}

// Bytes converts various data types to a byte slice for use in message payloads.
// This utility function handles common types used in IoT applications.
//
// Supported types:
//   - []byte: Returned as-is
//   - string: Converted to UTF-8 bytes
//   - int: Formatted as decimal string
//   - bool: Converted to "true" or "false"
//   - float64: Formatted with 2 decimal places (e.g., "25.50")
//
// Parameters:
//   - data: The value to convert to bytes
//
// Returns the byte representation and an error if the type is not supported.
//
// Example:
//
//	bytes, err := Bytes(25.5)  // Returns []byte("25.50"), nil
func Bytes(data any) ([]byte, error) {
	var buf []byte
	switch data.(type) {
	case []byte:
		buf = data.([]byte)

	case string:
		buf = []byte(data.(string))

	case int:
		str := fmt.Sprintf("%d", data.(int))
		buf = []byte(str)

	case bool:
		str := "false"
		if data.(bool) {
			str = "true"
		}
		buf = []byte(str)

	case float64:
		str := fmt.Sprintf("%5.2f", data.(float64))
		buf = []byte(str)

	default:
		err := errors.New("Can not convert data type " + fmt.Sprintf("%T", data))
		return nil, err
	}
	return buf, nil
}

// NewMsg creates a new message with the given topic, payload data, and source identifier.
// The topic is automatically parsed into path segments for routing, and the message
// is validated against Otto's topic format.
//
// If message saving is enabled via GetMsgSaver().StartSaving(), the message will
// be recorded for debugging purposes.
//
// Parameters:
//   - topic: The full topic path (e.g., "ss/c/station/temp")
//   - data: The message payload as bytes
//   - source: Identifier of the component creating this message (e.g., "mqtt-sub", "local")
//
// Returns a pointer to the newly created message.
//
// Example:
//
//	msg := NewMsg("ss/c/station1/temp", []byte("25.5"), "sensor")
//	fmt.Printf("Message ID: %d, Topic: %s\n", msg.ID, msg.Topic)
func NewMsg(topic string, data []byte, source string) *Msg {
	msg := &Msg{
		ID:        getMsgID(),
		Topic:     topic,
		Path:      strings.Split(topic, "/"),
		Data:      data,
		Timestamp: utils.Timestamp(),
		Source:    source,
		Valid:     ValidateTopic(topic),
	}

	if msgSaver != nil && msgSaver.Saving {
		msgSaver.Messages = append(msgSaver.Messages, msg)
	}
	return msg
}

// Station extracts the station identifier from the message topic.
// Otto topics follow the format "ss/[c|d]/station/sensor", where the
// third element (index 2) is the station ID.
//
// Returns the station ID string, or empty string if the topic doesn't
// have enough elements (less than 3).
//
// Example:
//
//	msg := NewMsg("ss/c/station1/temp", []byte("25"), "sensor")
//	station := msg.Station()  // Returns "station1"
func (msg *Msg) Station() string {
	if len(msg.Path) < 3 {
		return ""
	}
	return msg.Path[2]
}

// Last returns the last segment of the message topic path.
// This is typically the sensor or command name.
//
// Returns the last path element, or empty string if the path is empty.
//
// Example:
//
//	msg := NewMsg("ss/c/station1/temp", []byte("25"), "sensor")
//	last := msg.Last()  // Returns "temp"
func (msg *Msg) Last() string {
	l := len(msg.Path)
	if l == 0 {
		return ""
	}
	return msg.Path[l-1]
}

// Byte returns the raw byte array payload of the message.
// This is an alias for accessing the Data field directly.
func (msg *Msg) Byte() []byte {
	return msg.Data
}

// Bool interprets the message payload as a boolean value.
// It recognizes common boolean string representations:
//   - true: "true", "1", "on", "yes"
//   - false: everything else
//
// Returns true if the payload matches a truthy value, false otherwise.
//
// Example:
//
//	msg := NewMsg("ss/c/station/led", []byte("on"), "controller")
//	if msg.Bool() {
//	    fmt.Println("LED is on")
//	}
func (msg *Msg) Bool() bool {
	str := string(msg.Data)
	switch str {
	case "true", "1", "on", "yes":
		return true
	default:
		return false
	}
}

// String returns the message payload as a UTF-8 string.
// This is a convenience method for converting the byte payload to a string.
func (msg *Msg) String() string {
	return string(msg.Data)
}

// Float64 parses the message payload as a floating-point number.
// The payload should be a string representation of a number.
//
// Returns the parsed float64 value. If parsing fails, returns 0.0.
//
// Example:
//
//	msg := NewMsg("ss/d/station/temp", []byte("25.5"), "sensor")
//	temp := msg.Float64()  // Returns 25.5
func (msg *Msg) Float64() float64 {
	var f float64
	fmt.Sscanf(msg.String(), "%f", &f)
	return f
}

// IsJSON checks whether the message payload is valid JSON.
// This is useful for determining if the payload needs JSON parsing
// or should be treated as a plain string/binary data.
//
// Returns true if the payload is valid JSON, false otherwise.
func (msg *Msg) IsJSON() bool {
	return json.Valid(msg.Data)
}

// JSON encodes the entire message (not just the payload) as JSON.
// This serializes the Msg struct including all fields (ID, Topic, Data, etc.).
//
// Returns the JSON byte array and any encoding error.
//
// Note: To decode just the payload as JSON, use Map() instead.
func (msg *Msg) JSON() ([]byte, error) {
	jbytes, err := json.Marshal(msg)
	return jbytes, err
}

// Map decodes the message payload as a JSON object into a map.
// The payload must be valid JSON object format.
//
// Returns a map with string keys and interface{} values, or an error
// if the payload is not valid JSON or not an object.
//
// Example:
//
//	msg := NewMsg("ss/d/station/sensor", []byte(`{"temp":25.5,"humidity":60}`), "sensor")
//	data, err := msg.Map()
//	if err == nil {
//	    temp := data["temp"].(float64)
//	}
func (msg *Msg) Map() (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %s", err)
	}
	return m, nil
}

// Dump returns a human-readable multi-line string representation of the message
// including all fields. This is useful for debugging and logging.
//
// Returns a formatted string with ID, Path, Args, Source, Timestamp, and Data.
//
// Example output:
//
//	  ID: 123
//	Path: ["ss" "c" "station1" "temp"]
//	Args: []
//	 Src: sensor
//	Time: 1234567890
//	Data: 25.5
func (msg *Msg) Dump() string {
	str := fmt.Sprintf("  ID: %d\n", msg.ID)
	str += fmt.Sprintf("Path: %q\n", msg.Path)
	str += fmt.Sprintf("Args: %q\n", msg.Args)
	str += fmt.Sprintf(" Src: %s\n", msg.Source)
	str += fmt.Sprintf("Time: %s\n", msg.Timestamp)
	str += fmt.Sprintf("Data: %s\n", string(msg.Data))

	return str
}

// MsgSaver records messages for debugging, testing, and audit purposes.
// When saving is enabled, all messages created via NewMsg() are automatically
// appended to the Messages slice.
//
// This is useful for:
//   - Debugging message flow through the system
//   - Testing message handlers
//   - Auditing message history
//   - Replaying messages for testing
//
// Note: In production, be mindful of memory usage as messages accumulate.
// Consider periodically clearing or persisting messages to external storage.
//
// TODO: Add file persistence and protocol-based message delivery.
type MsgSaver struct {
	Messages []*Msg `json:"saved-messages"` // All recorded messages
	Saving   bool   `json:"saving"`         // Whether message recording is active
}

// GetMsgSaver returns the singleton MsgSaver instance, creating it if needed.
// This provides a global message recording facility that can be enabled or
// disabled at runtime.
//
// Returns a pointer to the MsgSaver instance.
//
// Example:
//
//	saver := GetMsgSaver()
//	saver.StartSaving()
//	// ... messages are now being recorded ...
//	saver.StopSaving()
func GetMsgSaver() *MsgSaver {
	if msgSaver == nil {
		msgSaver = &MsgSaver{}
	}
	return msgSaver
}

// StartSaving enables message recording. After calling this, all messages
// created via NewMsg() will be stored in the Messages slice.
//
// Example:
//
//	saver := GetMsgSaver()
//	saver.StartSaving()
func (ms *MsgSaver) StartSaving() {
	ms.Saving = true
}

// StopSaving disables message recording. Messages created after this call
// will not be saved. Previously saved messages remain in the Messages slice.
//
// Example:
//
//	saver := GetMsgSaver()
//	saver.StopSaving()
func (ms *MsgSaver) StopSaving() {
	ms.Saving = false
}

// Dump prints all saved messages to stdout in a human-readable format.
// Each message is separated by a line of dashes for readability.
//
// This is useful for debugging and inspecting the message history.
func (ms *MsgSaver) Dump() {
	for _, msg := range ms.Messages {
		println(msg.Dump())
		println("----------------------------------------------")
	}
}

// ServeHTTP implements http.Handler to provide a REST API endpoint for
// accessing saved messages. Returns the MsgSaver state as JSON, including
// all saved messages and the current saving status.
//
// Response format:
//
//	{
//	  "saved-messages": [{"id": 1, "topic": "...", ...}, ...],
//	  "saving": true
//	}
//
// This endpoint is useful for debugging and monitoring message flow.
func (ms *MsgSaver) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ms)
}
