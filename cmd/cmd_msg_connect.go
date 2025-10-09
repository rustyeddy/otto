/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rustyeddy/otto/messanger"
	"github.com/spf13/cobra"
)

// brokerCmd represents the broker command
var msgConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to the msg broker",
	Long:  `Connect to the msg broker`,
	Run:   runMsgConnect,
}

func runMsgConnect(cmd *cobra.Command, args []string) {
	messanger.GetMQTT()
}
