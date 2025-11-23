package commands

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
	"github.com/k4lipso/gokill/rpc"
)

///// REMOTE COMMANDS

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Interact with remote settings and state",
}

var remoteStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of one or more groups",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var reply []remote.PeerGroupInfo
		var ownPeerId string
		err := rpcClient.Call("Query.ListPeerGroups", 0, &reply)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		err = rpcClient.Call("Query.GetOwnPeerId", 0, &ownPeerId)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		var peerGroupId string
		if len(args) == 1 {
			peerGroupId = args[0]
		}

		var leveledList pterm.LeveledList

		colorize := func(state network.Connectedness, msg string) string {
			switch state {
			case network.NotConnected:
				return pterm.FgLightRed.Sprint(msg)
			case network.Connected:
				return pterm.FgLightGreen.Sprint(msg)
			case network.Limited:
				return pterm.FgLightRed.Sprint(msg)
			default:
				return msg
			}
		}

		peerGroupHeader := func(info remote.PeerGroupInfo) string {
			tmpStr := fmt.Sprintf("%s - %s", info.Name, info.Id)
			return tmpStr
		}

		for _, info := range reply {
			if len(peerGroupId) > 0 && peerGroupId != info.Id {
				continue
			}

			peerGroup := pterm.LeveledListItem{Level: 0, Text: peerGroupHeader(info)}
			leveledList = append(leveledList, peerGroup)

			if !showPeers {
				continue
			}

			peers := info.Peers

			if len(peers) < 1 {
				continue
			}

			for _, peer := range peers {
				if peer.Id == ownPeerId {
					continue
				}

				leveledList = append(leveledList, pterm.LeveledListItem{
					Level: 1,
					Text:  colorize(peer.ConnectionStatus, "●") + fmt.Sprintf(" %s Status: ", peer.Id) + colorize(peer.ConnectionStatus, peer.ConnectionStatus.String()),
				})
			}
		}
		root := putils.TreeFromLeveledList(leveledList)
		root.Text = "/"

		pterm.DefaultTree.WithRoot(root).Render()
	},
}

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
	Use:   "group",
	Short: "Add, delete or list groups",
}

var addPeerGroupCmd = &cobra.Command{
	Use:   "add",
	Short: "add a group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		peerGroup := args[0]

		var placeholder int
		err := rpcClient.Call("Query.AddPeerGroup", &peerGroup, &placeholder)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Infof("Group %s was added\n", peerGroup)
	},
}

var deletePeerGroupCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		peerGroup := args[0]

		var placeholder int
		err := rpcClient.Call("Query.DeletePeerGroup", &peerGroup, &placeholder)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		internal.Log.Infof("Group %s was deleted\n", peerGroup)
	},
}

var listPeerGroupsCmd = &cobra.Command{
	Use:   "list",
	Short: "list all groups",
	Run: func(cmd *cobra.Command, args []string) {
		var reply []remote.PeerGroupConfig
		err := rpcClient.Call("Query.ListPeerGroups", 0, &reply)

		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		for _, peerGroupCfg := range reply {
			internal.Log.Info(peerGroupCfg.Name)
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
