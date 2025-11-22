package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/k4lipso/gokill/rpc"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump gokill state",
	Run: func(cmd *cobra.Command, args []string) {
		var reply []rpc.TriggerInfo
		err := rpcClient.Call("Query.LoadedTriggers", 0, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		for _, info := range reply {
			fmt.Printf("TriggerName: %s\n", info.Config.Name)
			fmt.Printf("TriggerId: %s\n", info.Id.String())
			fmt.Printf("TriggerType: %s\n", info.Config.Type)
			fmt.Printf("TriggerIsActive: %v\n", info.Active)
			fmt.Printf("TriggerLoop: %v\n", info.Config.Loop)

			if !info.TimeStarted.IsZero() {
				fmt.Printf("TriggerRunningSince: %v seconds\n", time.Since(info.TimeStarted).Seconds())
			}

			if !info.TimeFired.IsZero() {
				fmt.Printf("TriggerFired %v seconds ago\n", time.Since(info.TimeFired).Seconds())
			}

			fmt.Printf("TriggerOptions: %s\n", info.Config.Options)

			for _, actions := range info.Config.Actions {
				fmt.Printf("TriggerActionType: %s\n", actions.Type)
				fmt.Printf("TriggerActionStage: %d\n", actions.Stage)
				fmt.Printf("TriggerActionOptions: %s\n", actions.Options)
			}
			fmt.Print("\n\n\n")
		}
	},
}
