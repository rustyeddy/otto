/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
)

// brokerCmd represents the broker command
var msgPubCmd = &cobra.Command{
	Use:   "pub",
	Short: "Publish to the msg topic",
	Long:  `Publish to msg tocpic`,
	Run:   runMsgPub,
}

func runMsgPub(cmd *cobra.Command, args []string) {
	m := messanger.GetMQTT()
	m.Publish(args[0], args[1])
}
