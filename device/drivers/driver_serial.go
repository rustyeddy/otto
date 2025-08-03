package drivers

import (
	"log/slog"

	"github.com/rustyeddy/otto/device"
	"go.bug.st/serial"
)

type Serial struct {
	PortName string
	Baud     int
	serial.Port
	mock bool
}

var (
	serialPorts map[string]*Serial
)

func init() {
	serialPorts = make(map[string]*Serial)
}

func GetSerial(port string) *Serial {
	if s, ex := serialPorts[port]; ex {
		return s
	}
	s, err := NewSerial(port, 115200)
	if err != nil {
		slog.Error("Serial port", "port", port, "error", err)
		return nil
	}
	return s
}

func NewSerial(port string, baud int) (s *Serial, err error) {
	s = &Serial{
		PortName: port,
		Baud:     baud,
	}

	if device.IsMock() {
		return s, nil
	}

	mode := &serial.Mode{
		BaudRate: baud,
	}
	s.Port, err = serial.Open(port, mode)
	if err != nil {
		return nil, err
	}
	serialPorts[port] = s
	return s, nil
}

// func NewSerial(name string, port string, opts any) *Device {
// 	sp := GetSerial(name, port, opts)
// 	err := sp.Open()
// 	return sp
// }

// type SerialDevice interface {
// 	Device
// 	Open() error
// 	PortName() string
// 	SerialReader
// 	SerialWriter
// }

// type SerialWriter interface {
// 	io.Reader
// }

// type SerialReader interface {
// 	io.Writer
// }

// type Closer interface {
// 	Close() error
// }

// func NewSerialDevice(name string, port string, opts any) SerialDevice {
// 	if mock {
// 		return GetSerialMock(name)
// 	}
// 	return GetSerialPort(name, port, opts)
// }

// type SerialMock struct {
// 	*BaseDevice
// 	portName string
// }

// func GetSerialMock(name string) *SerialMock {
// 	return &SerialMock{
// 		BaseDevice: NewDevice(name),
// 	}
// }

// func (d *SerialMock) Open() error {
// 	return nil
// }

// func (d *SerialMock) Read([]byte) (n int, err error) {
// 	return n, err
// }

// func (d *SerialMock) Write([]byte) (n int, err error) {
// 	return n, err
// }

// func (d *SerialMock) Close() error {
// 	return nil
// }

// func (d *SerialMock) PortName() string {
// 	return d.portName
// }
