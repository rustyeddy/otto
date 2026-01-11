package messenger

import "time"

// StatusPayload is the JSON body for status topics.
type StatusPayload struct {
	Status string    `json:"status"` // "online"|"offline"
	Time   time.Time `json:"time"`
}

// MetaPayload is the JSON body for device metadata topics.
type MetaPayload struct {
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`
	ValueType string            `json:"value_type"`
	Access    string            `json:"access"` // "ro"|"wo"|"rw"
	Unit      string            `json:"unit,omitempty"`
	Min       *float64          `json:"min,omitempty"`
	Max       *float64          `json:"max,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Attrs     map[string]string `json:"attrs,omitempty"`
}
