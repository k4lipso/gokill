package commands

import (
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"github.com/k4lipso/gokill/rpc"
	"github.com/k4lipso/gokill/triggers"
)

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
