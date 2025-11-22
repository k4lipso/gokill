package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

/// TRIGGERS

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Enable, disable, or test triggers",
}

var enableTriggerCmd = &cobra.Command{
	Use:   "enable",
	Short: "enable a trigger",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		triggerId := args[0]

		var reply *bool
		err := rpcClient.Call("Query.EnableTrigger", triggerId, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		fmt.Printf("%v", *reply)
	},
}

var disableTriggerCmd = &cobra.Command{
	Use:   "disable",
	Short: "disable a trigger",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		triggerId := args[0]

		var reply *bool
		err := rpcClient.Call("Query.DisableTrigger", triggerId, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		fmt.Printf("%v", *reply)
	},
}
