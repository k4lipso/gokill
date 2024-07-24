package main

import (
	"fmt"
	"time"
	"log"
	"flag"
	"github.com/k4lipso/gokill/rpc"
)

func main() {
	disableTrigger := flag.String("d", "", "Id of trigger you want to disable")
	flag.Parse()

	client, err := rpc.Receive()

	if err != nil {
		log.Fatal("dialing: ", err)
		return
	}

	if len(*disableTrigger) == 0 {
		var reply []rpc.TriggerInfo
		err = client.Call("Query.ActiveTriggers", 0, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		for _, info := range reply {
			fmt.Printf("TriggerName: %s\n", info.Config.Name)
			fmt.Printf("TriggerId: %s\n", info.Id.String())
			fmt.Printf("TriggerType: %s\n", info.Config.Type)

			if !info.TimeStarted.IsZero() {
				fmt.Printf("TriggerRunningSince: %v seconds\n", time.Now().Sub(info.TimeStarted).Seconds())
			}

			if !info.TimeFired.IsZero() {
				fmt.Printf("TriggerFired %v seconds ago\n", time.Now().Sub(info.TimeFired).Seconds())
			}
			fmt.Printf("TriggerType: %s\n", info.Config.Type)
			fmt.Printf("TriggerLoop: %v\n", info.Config.Loop)
			fmt.Printf("TriggerOptions: %s\n", info.Config.Options)

			for _, actions := range info.Config.Actions {
				fmt.Printf("TriggerActionType: %s\n", actions.Type)
				fmt.Printf("TriggerActionStage: %d\n", actions.Stage)
				fmt.Printf("TriggerActionOptions: %s\n", actions.Options)
			}
			fmt.Print("\n\n\n")
		}
	} else {
		var reply *bool
		err = client.Call("Query.DisableTrigger", disableTrigger, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		fmt.Printf("%v", *reply)
	}

}
