/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
)

// brokerCmd represents the broker command
var msgSubCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subscribe to the msg topic",
	Long:  `Subscribe to msg tocpic`,
	Run:   runMSGSub,
}

func runMSGSub(cmd *cobra.Command, args []string) {
	m := messanger.GetMQTT()
	if m.Client == nil || !m.IsConnected() {
		fmt.Fprintf(cmdOutput, "Warning MQTT is not connected to mqtt broker: %s\n", m.Broker)
	}

	p := &messanger.MsgPrinter{}
	m.Subscribe(args[0], p.MsgHandler)
}
