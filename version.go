package otto

import "fmt"

var Version = "0.0.12"

func VersionJSON() []byte {
	return []byte(fmt.Sprintf(`{"version": "%s"}`, Version))
}
