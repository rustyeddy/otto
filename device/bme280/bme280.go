package bme280

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/maciej/bme280"
	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/device/drivers"
)

// BME280 is an I2C temperature, humidity and pressure sensor. It
// defaults to address 0x77
type BME280 struct {
	*device.Device

	bus    string
	addr   int
	driver *bme280.Driver
}

type Env struct {
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
	Pressure    string `json:"pressure"`
}

// Response returns values read from the sensor containing all three
// values for temperature, humidity and pressure
type Response bme280.Response

// Create a new BME280 at the give bus and address. Defaults are
// typically /dev/i2c-1 address 0x99
func New(name, bus string, addr int) *BME280 {
	b := &BME280{
		Device: device.NewDevice(name),
		bus:    bus,
		addr:   addr,
	}
	return b
}

// Init opens the i2c bus at the specified address and gets the device
// ready for reading
func (b *BME280) Init() error {
	if device.IsMock() == true {
		return nil
	}

	i2c, err := drivers.GetI2CDriver(b.bus, b.addr)
	if err != nil {
		return err
	}

	b.driver = bme280.New(i2c)
	err = b.driver.InitWith(bme280.ModeForced, bme280.Settings{
		Filter:                  bme280.FilterOff,
		Standby:                 bme280.StandByTime1000ms,
		PressureOversampling:    bme280.Oversampling16x,
		TemperatureOversampling: bme280.Oversampling16x,
		HumidityOversampling:    bme280.Oversampling16x,
	})
	if err != nil {
		return err
	}
	return nil
}

// Read one Response from the sensor. If this device is being mocked
// we will make up some random floating point numbers between 0 and
// 100.
func (b *BME280) Read() (*bme280.Response, error) {
	if device.IsMock() {
		return &bme280.Response{
			Temperature: rand.Float64() * 100,
			Pressure:    rand.Float64() * 100,
			Humidity:    rand.Float64() * 100,
		}, nil
	}

	response, err := b.driver.Read()
	if err != nil {
		return nil, err
	}
	return &response, err
}

// ReadPub reads the latest values from the sendsor then publishes
// them on the MQTT topic assigned to this device.
func (b *BME280) ReadPub() error {
	vals, err := b.Read()
	if err != nil {
		return errors.New("Failed to read bme280: " + err.Error())
	}

	vals.Temperature = (vals.Temperature * (9 / 5)) + 32

	valstr := &Env{
		Temperature: fmt.Sprintf("%.2f", vals.Temperature),
		Humidity:    fmt.Sprintf("%.2f", vals.Humidity),
		Pressure:    fmt.Sprintf("%.2f", vals.Pressure),
	}

	jb, err := json.Marshal(valstr)
	if err != nil {
		return errors.New("BME280 failed marshal read response" + err.Error())
	}
	b.PubData(jb)
	return nil
}
