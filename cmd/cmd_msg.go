/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
)

type msgConfiguration struct {
	Broker  string
	Enabled bool
}

var (
	msgConfig msgConfiguration

	// msgCmd represents the msg command
	msgCmd = &cobra.Command{
		Use:   "msg",
		Short: "Configure and interact with MSG broker",
		Long:  `This command can be used to interact and diagnose an MSG broker`,
		Run:   msgRun,
	}

	cmdWriter io.Writer = os.Stdout
)

func init() {
	msgCmd.PersistentFlags().StringVar(&msgConfig.Broker, "broker", "localhost", "Set the MSG Broker")
	cmdWriter = io.Discard
}

func msgRun(cmd *cobra.Command, args []string) {
	m := messanger.GetMQTT()

	// If the broker config changes and msg is connected, disconnect
	// and reconnect to new broker
	if msgConfig.Broker != m.Broker {
		m.Broker = msgConfig.Broker
	}

	connected := false
	if m.Client != nil {
		connected = m.IsConnected()
	}

	fmt.Fprintf(cmdWriter, "Broker: %s\n", m.Broker)
	fmt.Fprintf(cmdWriter, "Connected: %t\n", connected)
	fmt.Fprintf(cmdWriter, "Debug: %t\n", m.Debug)
	fmt.Fprintln(cmdWriter, "\nSubscriptions")
}
