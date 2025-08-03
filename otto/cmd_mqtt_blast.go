/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package otto

import (
	"github.com/rustyeddy/otto/blasters"

	"github.com/spf13/cobra"
)

var (
	count int

	// mqttCmd represents the mqtt command
	mqttBlastCmd = &cobra.Command{
		Use:   "blast",
		Short: "Start blasting MQTT messages from (count) blasters",
		Long:  `Start blasting MQTT messages from (count) blasters`,
		Run:   mqttBlastRun,
	}
)

func init() {
	mqttCmd.AddCommand(mqttBlastCmd)
	mqttBlastCmd.PersistentFlags().IntVar(&count, "count", 1, "The number of blasters to start")
}

func mqttBlastRun(cmd *cobra.Command, args []string) {
	blasters := blasters.NewMQTTBlasters(count)
	go blasters.Blast()
}
