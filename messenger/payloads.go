package messenger

import "time"

type StatusPayload struct {
	Status string    `json:"status"` // "online"|"offline"
	Time   time.Time `json:"time"`
}

type MetaPayload struct {
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`
	ValueType string            `json:"value_type"`
	Access    string            `json:"access"` // "ro"|"wo"|"rw"
	Unit      string            `json:"unit,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Attrs     map[string]string `json:"attrs,omitempty"`
}
