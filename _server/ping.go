package server

import (
	"fmt"
	"net/http"
)

// Ping is a full fledged Request handler, You can write your own!
type Ping struct {
}

// ServeHTTP will respond to the writer with 'Pong'
func (p Ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Pong\n")
}
