package station

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tstdevice struct {
	name string
}

func (da *tstdevice) Name() string {
	return da.name
}

func TestDeviceManager(t *testing.T) {
	assert.Nil(t, deviceManager)
	dm := GetDeviceManager()
	assert.NotNil(t, dm)
	assert.NotNil(t, deviceManager)
}

func TestDeviceManagerAdd(t *testing.T) {
	dm := GetDeviceManager()
	md := dm.Add(&tstdevice{name: "device"})
	assert.NotNil(t, md)

	t.Run("NilRegister", func(t *testing.T) {
		dm.Register(nil)
	})

	t.Run("Get", func(t *testing.T) {
		d := dm.Get("not-registered")
		assert.Nil(t, d)

		dev := dm.GetDevice("not-registered")
		assert.Nil(t, dev)

		d = dm.Get("device")
		assert.Equal(t, md, d)

		dev = dm.GetDevice("device")
		assert.NotNil(t, dev)
		assert.Equal(t, d.Device, dev)
	})

	t.Run("Remove", func(t *testing.T) {
		assert.Equal(t, 1, len(dm.devices))
		dm.Remove("nothing")
		assert.Equal(t, 1, len(dm.devices))

		dm.Remove("device")
		assert.Equal(t, 0, len(dm.devices))
	})
}
