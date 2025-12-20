/*
OttO is a set of Go packages (framework) that help build IoT
applications. The goal is to decouple the IoT sensors, actuators,
etc. from the framework hardware interfaces such as GPIO, I2C, serial
ports, etc.

# Features include

Device level abstraction. Each device has a name that translates into
a path that can be used by MQTT and HTTP REST interface for
communication with other systems.

Device manager that keeps track of all application devices,
configuration and status. This interface attempts to be agnostic to
the underlying drivers, including gpiocdev, I2C, periph.io, etc.

Message based architecture abstracting all communications into a
standard message format. With functionality that can save messages for
later replay or diagnostics.

MQTT messaging built into all devices and components according to
functionality and need

HTTP Rest interface and corresponding API for all components of the
framework.

Drivers for a few different breakout boards meant to run on the
Raspberry Pi.

Station module to represent a single application on a given device or
a series of stations for a networked controller.

Messanger (not to be confused with messages) implements a Pub/Sub
(MQTT or other) interface between components of your application

# HTTP REST Server for data gathering and configuration

# Websockets for realtime bidirectional communication with a UI

High performance Web server built in to serve interactive UI's
and modern API's

Station manager to manage the stations that make up an entire sensor
network

Data Manager for temporary data caching and interfaces to update
your favorite cloud based timeseries database

Message library for standardized messages built to be communicate
events and information between pacakges.

The primary communication model for OttO is a messaging system based
on the Pub/Sub model defaulting to MQTT. oTTo is also heavily invested
in HTTP to implement user interfaces and REST/Graph APIs.

Messaging and HTTP use paths to specify the subject of interest. These
paths can be generically reduced to an ordered collection of strings
seperated by slashes '/'.  Both MQTT topics, http URI's and UNIX
filesystems use this same schema which we use the generalize the
identity of the elements we are addressing.

In other words we can generalize the following identities:

For example:

	    File: /home/rusty/data/hb/temperature
		HTTP: /api/data/hb/temperature
		MQTT: ss/station/hb/temperature

The data within the highest level topic temperature can be represented
say by JSON `{ farenhiet: 100.07 }`

### Meta Data (Station Information)

For example when a station comes alive it can provide some information
about itself using the topic:

	```ss/m/be:ef:ca:fe:02/station```

The station will announce itself along with some meta information and
it's capabilities.  The body of the message might look something like
this:

```json

	{
		"id": "be:ef:ca:fe:02",
		"ip": "10.11.24.24",
	    "sensors": [
			"tempc",
			"humidity",
			"light"
		],
		"relays": [
			"heater",
			"light"
		],
	}

```

### Sensor Data

Sensor data takes on the form:

	```ss/d/<station>/<sensor>/<index>```

Where the source is the Station ID publishing the respective data.
The sensor is the type of data being produced (temp, humidity,
lidar, GPS).

The index is useful in application where there is more than one
device, such as sensors, motors, etc.

The value published by the sensors is typically going to be floating
point, however these values may also be integers, strings or byte
arrays.

### Control Data

	```ss/c/<source>/<device>/<index>```

This is essentially the same as the sensor except that control
commands are used to have a particular device change, for example
turning a relay on or off.
*/
package otto

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/server"
	"github.com/rustyeddy/otto/station"
)

// Controller is a message handler that oversees all interactions
// with the application.
type Controller interface {
	Init()
	Start() error
	Stop()
	MsgHandler(m *messanger.Msg)
}

// OttO is a large wrapper around the Station, Server,
// DataManager and Messanger, including some convenience functions.
type OttO struct {
	Name string

	*station.Station
	*station.StationManager
	*server.Server
	messanger.Messanger

	Mock       bool
	MQTTBroker string // MQTT broker URL, defaults to test.mosquitto.org
	UseLocal   bool   // Force use of local messaging
	hub        bool   // maybe hub should be a different struct?
	done       chan any
}

// global variables and structures
var (
	Version     string
	Interactive bool
)

func init() {
	Version = "0.0.9"
}

func (o *OttO) Done() chan any {
	return o.done
}

// OttO is a convinience function starting the MQTT and HTTP servers,
// the station manager and other stuff.
func (o *OttO) Init() {
	if o.done != nil {
		// server has already been started
		fmt.Println("Server has already been started")
		return
	}
	o.done = make(chan any)

	if o.StationManager != nil || o.Station != nil || o.Messanger != nil {
		str := "OttO Init has been called twice, one of these is not nil\n" +
			fmt.Sprintf("\tStationManager (%p)\n", o.StationManager) +
			fmt.Sprintf("\tStation (%p)\n", o.Station) +
			fmt.Sprintf("\tServer (%p)\n", o.Server) +
			fmt.Sprintf("\tMessanger (%p)\n", o.Messanger)
		panic(str)
	}

	var err error
	o.StationManager = station.GetStationManager()
	o.Server = server.GetServer()
    o.Name = "myname"
	o.Station, err = o.StationManager.Add(o.Name)
	if err != nil {
		slog.Error("Unable to create station", "error", err)
		return
	}
	// Initialzie the local station
	o.Station.Init()
	o.Messanger = messanger.GetMessanger()
}

// Start the OttO process, TODO return a stop channel or context?
func (o *OttO) Start() {
	if o.Messanger != nil {
		go o.Messanger.Connect()
	}

	if o.Server != nil {
		go o.Server.Start(o.done)
	}

	if o.StationManager != nil {
		go o.StationManager.Start()
	}
}

func (o *OttO) Stop() {
	o.done <- true
	slog.Info("Done, cleaning up()")

	if err := server.GetServer().Close(); err != nil {
		slog.Error("Failed to close server", "error", err)
	}

	if o.Messanger != nil {
		o.Messanger.Close()
		messanger.StopMQTTBroker(context.Background())
	}
}

// AddManagedDevice creates a managed device wrapper and adds it to the station
func (o *OttO) AddManagedDevice(name string, device any, topic string) *station.ManagedDevice {
	md := station.NewManagedDevice(name, device, topic)
	if o.Station != nil {
		o.Station.Register(md)
	}
	return md
}

// GetManagedDevice retrieves a managed device by name
func (o *OttO) GetManagedDevice(name string) *station.ManagedDevice {
	if o.Station == nil {
		return nil
	}
	device := o.Station.Get(name)
	return device
}
