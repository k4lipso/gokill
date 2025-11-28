package commands

import (
	"github.com/spf13/cobra"

	RPC "net/rpc"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/rpc"
)

var (
	rpcClient  *RPC.Client
	dbPath     string
	debug      bool
	showStages bool
	showPeers  bool
)

// Create the root command
var RootCmd = &cobra.Command{
	Use:   "gokillctl",
	Short: "Interact with the gokill daemon",
}

// /// REMOTE COMMANDS
func init() {
	cobra.OnInitialize(func() {
		internal.InitLogger()
		internal.SetVerbose(debug)

		var tmpClient *RPC.Client
		tmpClient, err := rpc.Receive(dbPath)
		rpcClient = tmpClient

		if err != nil {
			internal.Log.Fatalf("dialing: %s\n", err)
			return
		}
	})

	RootCmd.PersistentFlags().StringVar(&dbPath, "db", "./db", "db path")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	statusCmd.Flags().BoolVar(&showStages, "stages", false, "Show configured Stages")
	remoteStatusCmd.Flags().BoolVar(&showPeers, "peers", false, "Show peering status")
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(dumpCmd)

	triggerCmd.AddCommand(enableTriggerCmd)
	triggerCmd.AddCommand(disableTriggerCmd)
	RootCmd.AddCommand(triggerCmd)

	peerCmd.AddCommand(addPeerCmd)
	peerCmd.AddCommand(removePeerCmd)
	peerCmd.AddCommand(infoCmd)

	peerGroupCmd.AddCommand(addPeerGroupCmd)
	peerGroupCmd.AddCommand(deletePeerGroupCmd)
	peerGroupCmd.AddCommand(listPeerGroupsCmd)
	peerGroupCmd.AddCommand(getPeerGroupIdCmd)
	peerGroupCmd.AddCommand(updatePeerGroupIdCmd)

	remoteCmd.AddCommand(remoteStatusCmd)
	remoteCmd.AddCommand(broadcastCmd)
	remoteCmd.AddCommand(peerCmd)
	remoteCmd.AddCommand(peerGroupCmd)
	RootCmd.AddCommand(remoteCmd)
}
