package messanger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/rustyeddy/otto/utils"
)

// Topics maintains the list of topics used by otto and the
// applications. It maintains the topic format and a count for each
// time the topic is used
type Topics struct {
	TopicFmt string
	Topicmap map[string]int
}

var (
	topics *Topics
)

func init() {
	topics = &Topics{
		TopicFmt: "ss/%s/%s/%s",
		Topicmap: make(map[string]int),
	}
}

// validate topic
func ValidateTopic(topic string) bool {
	path := strings.Split(topic, "/")
	if len(path) < 4 {
		return false
	}

	if path[0] != "ss" {
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

// GetTopics will return the Topics structure, one per application.
func GetTopics() *Topics {
	return topics
}

// Control will return a control topic e.g. ss/c/station/foo
func (t *Topics) Control(topic string) string {
	top := fmt.Sprintf(t.TopicFmt, "c", utils.StationName(), topic)
	t.Topicmap[top]++
	return top
}

// Control will return a data topic e.g. ss/d/station/foo
func (t *Topics) Data(topic string) string {
	top := fmt.Sprintf(t.TopicFmt, "d", utils.StationName(), topic)
	t.Topicmap[top]++
	return top
}

// ServeHTTP is a JSON endpoint that returns all the topics used by
// this station.
func (t Topics) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(t)
	if err != nil {
		slog.Error("Error wrote data", "error", err)
	}
}
