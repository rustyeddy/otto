package messanger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMQTTSubscribeAndConnect(t *testing.T) {

	m, err := NewMessangerMQTT("testid", "localhost")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	var sub1, sub2, sub3 []byte
	m.Subscribe("subscription1", func(msg *Msg) error {
		sub1 = []byte("sub1")
		return nil
	})
	m.Subscribe("subscription2", func(msg *Msg) error {
		sub2 = []byte("sub2")
		return nil
	})
	m.Subscribe("subscription3", func(msg *Msg) error {
		sub3 = []byte("sub3")
		return nil
	})

	assert.NoError(t, m.Connect())
	msg := NewMsg("subscription1", []byte(sub1), "test")
	m.PubMsg(msg)
	assert.Equal(t, msg.Data, sub1)

	msg = NewMsg("subscription2", []byte(sub2), "test")
	m.PubMsg(msg)
	assert.Equal(t, msg.Data, sub2)

	msg = NewMsg("subscription3", []byte(sub3), "test")
	m.PubMsg(msg)
	assert.Equal(t, msg.Data, sub2)

	m.Close()
	assert.False(t, m.IsConnected())
	assert.Equal(t, len(m.subs), 0)

}
