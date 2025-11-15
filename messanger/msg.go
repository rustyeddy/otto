package messanger

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rustyeddy/otto/utils"
)

// Msg holds a value and some type of meta data to be pass around in
// the system. The Msg struct contains all the info need to communicate
// internally or over the PubSub protocol.  Every Msg has a unique ID
// that can optionally be tracked, saved or replayed for debugging or
// testing purposes.
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
	msgid    int64 = 0
	msgSaver *MsgSaver
)

// getMsgID returns a globally unique message ID. It simply increments
// the ID by 1 every time it is called. This ID will uniquely identify
// exact elements used by the system.
func getMsgID() int64 {
	msgid++
	return msgid
}

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

// New creates a new Msg from the given topic, data and a source
// string.
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

// Station extracts the station element from the Msg topic and returns
// the station ID/name to the caller.
func (msg *Msg) Station() string {
	if len(msg.Path) < 3 {
		return ""
	}
	return msg.Path[2]
}

// Last returns the Last element in the Msg.Topic path
func (msg *Msg) Last() string {
	l := len(msg.Path)
	if l == 0 {
		return ""
	}
	return msg.Path[l-1]
}

// Byte returns the array version of the Msg.Data
func (msg *Msg) Byte() []byte {
	return msg.Data
}

func (msg *Msg) Bool() bool {
	str := string(msg.Data)
	switch str {
	case "true", "1", "on", "yes":
		return true
	default:
		return false
	}
}

// String returns the string formatted version of Msg.Data
func (msg *Msg) String() string {
	return string(msg.Data)
}

// Float64 returns the float64 version of the Msg.Data
func (msg *Msg) Float64() float64 {
	var f float64
	fmt.Sscanf(msg.String(), "%f", &f)
	return f
}

// IsJSON returns true or false to indicate if the Msg.Data payload is
// a JSON formatted string/byte array or not.
func (msg *Msg) IsJSON() bool {
	return json.Valid(msg.Data)
}

// JSON encodes the Msg.Data into a JSON formatted byte array.
func (msg *Msg) JSON() ([]byte, error) {
	jbytes, err := json.Marshal(msg)
	return jbytes, err
}

// Map decodes the Msg.Data payload from a JSON formatted byte array
// into a map where the key/value pairs are the data index and values.
func (msg *Msg) Map() (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %s", err)
	}
	return m, nil
}

// Dump spits out the fields and values of the Msg data struct
func (msg *Msg) Dump() string {
	str := fmt.Sprintf("  ID: %d\n", msg.ID)
	str += fmt.Sprintf("Path: %q\n", msg.Path)
	str += fmt.Sprintf("Args: %q\n", msg.Args)
	str += fmt.Sprintf(" Src: %s\n", msg.Source)
	str += fmt.Sprintf("Time: %s\n", msg.Timestamp)
	str += fmt.Sprintf("Data: %s\n", string(msg.Data))

	return str
}

// MsgSaver struct is used to store a historical record of message
// captured by the application. Save the messages can be turned on and
// off at any given time.  TODO: need to be able to save these
// messages to a file, or deliver them via a protocol.
type MsgSaver struct {
	Messages []*Msg `json:"saved-messages"`
	Saving   bool   `json:"saving"`
}

// GetMsgSaver will return the instance of the MsgSaver element. The
// first time this funcction is called the object will be created.
func GetMsgSaver() *MsgSaver {
	if msgSaver == nil {
		msgSaver = &MsgSaver{}
	}
	return msgSaver
}

// StartSaving turn on message saving
func (ms *MsgSaver) StartSaving() {
	ms.Saving = true
}

// StopSaving disable message saving
func (ms *MsgSaver) StopSaving() {
	ms.Saving = false
}

// Dump spits out the history of messages
func (ms *MsgSaver) Dump() {
	for _, msg := range ms.Messages {
		println(msg.Dump())
		println("----------------------------------------------")
	}
}

// ServeHTTP will respond to the writer with 'Pong'
func (ms *MsgSaver) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ms)
}
