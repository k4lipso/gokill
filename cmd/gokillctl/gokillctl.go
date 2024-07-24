package main

import (
	"fmt"
	"log"
	"flag"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/k4lipso/gokill/rpc"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func main() {
	disableTrigger := flag.String("d", "", "Id of trigger you want to disable")
	enableTrigger := flag.String("e", "", "Id of trigger you want to enable")
	flag.Parse()

	client, err := rpc.Receive()

	if err != nil {
		log.Fatal("dialing: ", err)
		return
	}

	if len(*disableTrigger) == 0 && len(*enableTrigger) == 0 {

		var reply []rpc.TriggerInfo
		err = client.Call("Query.LoadedTriggers", 0, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		var items []list.Item

		for _, info := range reply {
			items = append(items, info)
		}


		m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
		m.list.Title = "Loaded Triggers"

		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			return
		}

		//for _, info := range reply {
		//	fmt.Printf("TriggerName: %s\n", info.Config.Name)
		//	fmt.Printf("TriggerId: %s\n", info.Id.String())
		//	fmt.Printf("TriggerType: %s\n", info.Config.Type)
		//	fmt.Printf("TriggerIsActive: %v\n", info.Active)
		//	fmt.Printf("TriggerLoop: %v\n", info.Config.Loop)

		//	if !info.TimeStarted.IsZero() {
		//		fmt.Printf("TriggerRunningSince: %v seconds\n", time.Now().Sub(info.TimeStarted).Seconds())
		//	}

		//	if !info.TimeFired.IsZero() {
		//		fmt.Printf("TriggerFired %v seconds ago\n", time.Now().Sub(info.TimeFired).Seconds())
		//	}

		//	fmt.Printf("TriggerOptions: %s\n", info.Config.Options)

		//	for _, actions := range info.Config.Actions {
		//		fmt.Printf("TriggerActionType: %s\n", actions.Type)
		//		fmt.Printf("TriggerActionStage: %d\n", actions.Stage)
		//		fmt.Printf("TriggerActionOptions: %s\n", actions.Options)
		//	}
		//	fmt.Print("\n\n\n")
		//}
	} 

	if len(*disableTrigger) != 0 {
		var reply *bool
		err = client.Call("Query.DisableTrigger", disableTrigger, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		fmt.Printf("%v", *reply)
	}

	if len(*enableTrigger) != 0 {
		var reply *bool
		err = client.Call("Query.EnableTrigger", enableTrigger, &reply)

		if err != nil {
			log.Fatal("query error:", err)
		}

		fmt.Printf("%v", *reply)
	}

}
