module github.com/rustyeddy/otto

go 1.24.5

require (
	github.com/chzyer/readline v1.5.1
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/gorilla/websocket v1.5.3
	github.com/mochi-mqtt/server/v2 v2.7.9
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.11.1
)

replace github.com/rustyeddy/devices => ../devices

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/rustyeddy/devices v0.0.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
