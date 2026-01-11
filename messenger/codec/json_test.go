package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	c := JSON[int]{}

	raw, err := c.Marshal(42)
	require.NoError(t, err)

	got, err := c.Unmarshal(raw)
	require.NoError(t, err)
	assert.Equal(t, 42, got)
}

func TestJSONUnmarshalInvalid(t *testing.T) {
	t.Parallel()

	c := JSON[int]{}
	_, err := c.Unmarshal([]byte(`"not-an-int"`))
	require.Error(t, err)
}
