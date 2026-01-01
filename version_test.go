package otto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionJSON(t *testing.T) {
	vbytes := VersionJSON()

	vmap := make(map[string]string)
	err := json.Unmarshal(vbytes, &vmap)
	assert.NoError(t, err)
	assert.Equal(t, Version, vmap["version"])
}
