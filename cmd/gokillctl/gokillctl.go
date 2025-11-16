package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/rpc"
	"github.com/k4lipso/gokill/triggers"
	RPC "net/rpc"
)

var (
	dbPath     string
	debug      bool
	showStages bool
	rpcClient  *RPC.Client
)

// Create the root command
var rootCmd = &cobra.Command{
	Use:   "gokillctl",
	Short: "Interact with the gokill daemon",
}

///// REMOTE COMMANDS

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Interact with remote settings and state",
}

// TODO: Continue here, create 'write' command to write something to the topic and test if its received
var broadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "Broadcast a message to root",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		peerGroup := rpc.PeerGroupService{PeerGroup: "root", Service: serviceName}
		err := rpcClient.Call("Query.Broadcast", &peerGroup, "")
		if err != nil {
			internal.Log.Error(err.Error())
		}

		internal.Log.Info("Broadcasted message.")
	},
}

var peerGroupCmd = &cobra.Command{
	Use:   "peerGroup",
	Short: "Add, delete or list peerGroups",
}

var addPeerGroupCmd = &cobra.Command{
	Use:   "add",
	Short: "add a peerGroup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		peerGroup := args[0]

		var placeholder int
		err := rpcClient.Call("Query.AddPeerGroup", &peerGroup, &placeholder)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Infof("PeerGroup %s was added\n", peerGroup)
	},
}

var deletePeerGroupCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a peerGroup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		peerGroup := args[0]

		var placeholder int
		err := rpcClient.Call("Query.DeletePeerGroup", &peerGroup, &placeholder)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Infof("PeerGroup %s was deleted\n", peerGroup)
	},
}

var listPeerGroupsCmd = &cobra.Command{
	Use:   "list",
	Short: "list all peerGroups",
	Run: func(cmd *cobra.Command, args []string) {
		var reply []string
		err := rpcClient.Call("Query.ListPeerGroups", 0, &reply)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Info("PeerGroups:")
		for _, ns := range reply {
			internal.Log.Info(ns)
		}
	},
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Add or remove peers. Get your own peering information",
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "print your own peerstring",
	Run: func(cmd *cobra.Command, args []string) {
		var result *string
		err := rpcClient.Call("Query.GetPeerString", 0, &result)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Info(*result)
	},
}

var addPeerCmd = &cobra.Command{
	Use:   "add",
	Short: "add a peer",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var peerGroup string
		var peerString string

		if len(args) == 1 {
			peerGroup = "root"
			peerString = args[0]
		} else {
			peerGroup = args[0]
			peerString = args[1]
		}

		var success *bool
		np := rpc.PeerGroupPeer{PeerGroup: peerGroup, Peer: peerString}
		err := rpcClient.Call("Query.AddPeer", &np, &success)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		if *success {
			internal.Log.Infof("Added peer: %s", peerString)
		} else {
			internal.Log.Infof("Could not add peer: %s", peerString)
		}
	},
}

var removePeerCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a peer",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var peerGroup string
		var peerString string

		if len(args) == 1 {
			peerGroup = "root"
			peerString = args[0]
		} else {
			peerGroup = args[0]
			peerString = args[1]
		}

		var success *bool
		np := rpc.PeerGroupPeer{PeerGroup: peerGroup, Peer: peerString}
		err := rpcClient.Call("Query.DeletePeer", &np, &success)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		if *success {
			internal.Log.Infof("Removed peer: %s", peerString)
		} else {
			internal.Log.Infof("Could not find peer: %s", peerString)
		}
	},
}

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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show runtime status of on or more triggers",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var reply []rpc.TriggerInfo
		err := rpcClient.Call("Query.LoadedTriggers", 0, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		var triggerId string

		if len(args) == 1 {
			triggerId = args[0]
		}

		var leveledList pterm.LeveledList

		colorize := func(state triggers.TriggerState, msg string) string {
			switch state {
			case triggers.Initialized:
				return msg
			case triggers.Test:
				return pterm.FgLightBlue.Sprint(msg)
			case triggers.Armed:
				return pterm.FgLightGreen.Sprint(msg)
			case triggers.Firing:
				return pterm.FgLightRed.Sprint(msg)
			case triggers.Failed:
				return pterm.FgRed.Sprint(msg)
			case triggers.Done:
				return pterm.FgLightWhite.Sprint(msg)
			case triggers.Disabled:
				return msg
			default:
				return msg
			}
		}

		triggerHeader := func(info rpc.TriggerInfo) string {
			tmpStr := fmt.Sprintf(" TRIGGER: %s - %s, Type: %s ", info.Id.String(), info.Config.Name, info.Config.Type)
			return colorize(info.State, "●") + tmpStr + colorize(info.State, "Status: "+info.State.String())
		}

		for _, info := range reply {
			if len(triggerId) > 0 && triggerId != info.Id.String() {
				continue
			}

			trigger := pterm.LeveledListItem{Level: 0, Text: triggerHeader(info)}
			leveledList = append(leveledList, trigger)

			if !showStages {
				continue
			}

			actionsCfg := info.Config.Actions

			if len(actionsCfg) < 1 {
				continue
			}

			sort.Slice(actionsCfg, func(i, j int) bool {
				return actionsCfg[i].Stage < actionsCfg[j].Stage
			})

			currentStage := actionsCfg[0].Stage

			writeCurrentStage := func() {
				leveledList = append(leveledList, pterm.LeveledListItem{
					Level: 1,
					Text:  fmt.Sprintf("STAGE %d", currentStage),
				})
			}

			writeCurrentStage()

			for _, action := range info.Config.Actions {
				if action.Stage != currentStage {
					currentStage = action.Stage
					writeCurrentStage()
				}

				leveledList = append(leveledList, pterm.LeveledListItem{
					Level: 2,
					Text:  fmt.Sprintf("ACTION: %s", action.Type),
				})
			}
		}
		root := putils.TreeFromLeveledList(leveledList)
		root.Text = "/"

		pterm.DefaultTree.WithRoot(root).Render()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./db", "db path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	statusCmd.Flags().BoolVar(&showStages, "stages", false, "Show configured Stages")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(dumpCmd)

	triggerCmd.AddCommand(enableTriggerCmd)
	triggerCmd.AddCommand(disableTriggerCmd)
	rootCmd.AddCommand(triggerCmd)

	peerCmd.AddCommand(addPeerCmd)
	peerCmd.AddCommand(removePeerCmd)
	peerCmd.AddCommand(infoCmd)

	peerGroupCmd.AddCommand(addPeerGroupCmd)
	peerGroupCmd.AddCommand(deletePeerGroupCmd)
	peerGroupCmd.AddCommand(listPeerGroupsCmd)

	remoteCmd.AddCommand(broadcastCmd)
	remoteCmd.AddCommand(peerCmd)
	remoteCmd.AddCommand(peerGroupCmd)
	rootCmd.AddCommand(remoteCmd)
}

func main() {
	cobra.OnInitialize(func() {
		var tmpClient *RPC.Client
		tmpClient, err := rpc.Receive(dbPath)
		rpcClient = tmpClient

		if err != nil {
			internal.Log.Fatalf("dialing: %s\n", err)
			return
		}

		internal.InitLogger()
		internal.SetVerbose(debug)
	})

	if err := rootCmd.Execute(); err != nil {
		internal.Log.Error(err.Error())
		os.Exit(1)
	}
}
