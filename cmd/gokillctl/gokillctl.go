package main

import (
	"fmt"
	"log"
	"flag"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	rpcExt "net/rpc"
	"github.com/k4lipso/gokill/rpc"
)

var client *rpcExt.Client

var (
	modelStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#555555"))

	focusedModelStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#ffffff"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#aa0000"))
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00aa00"))
)

type mainModel struct {
	state   int
	infos   []rpc.TriggerInfo
	index   int
	WindowSize tea.WindowSizeMsg
	Selected int
}

func newModel(infos []rpc.TriggerInfo) mainModel {
	m := mainModel{state: 0, Selected: -1, infos: infos,}

	return m
}

func (m mainModel) Render() string {

	var horizontal []string
	var vertical [][]string

	cellsPerRow := 3
	cellWidth := (m.WindowSize.Width / cellsPerRow) - int(0.02 * float32(m.WindowSize.Width))
	cellHeight := m.WindowSize.Height / (int(len(m.infos) / cellsPerRow) + 3)
	cellWidth = 15
	cellHeight = 5

	style := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(modelStyle)
	focusedStyle := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(focusedModelStyle)

	count := 0
	for idx, info := range m.infos {
		activeStyled := activeStyle.Render("Active")
		if info.Active == false {
			activeStyled = disabledStyle.Render("Disabled")
		}

		if m.state == idx {
			horizontal = append(horizontal, focusedStyle.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.Title(), info.Description(), activeStyled)))
		} else {
			horizontal = append(horizontal, style.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.Title(), info.Description(), activeStyled)))
		}

		count += 1
		if count == cellsPerRow || (count < cellsPerRow && idx == len(m.infos) - 1) {
			vertical = append(vertical, horizontal)
			horizontal = []string{}
			count = 0
		}
	}

	
	var horizontalRendered []string
	for _, v := range vertical {
		horizontalRendered = append(horizontalRendered, lipgloss.JoinHorizontal(lipgloss.Top, v...))
	}

	return lipgloss.JoinVertical(lipgloss.Top, horizontalRendered...)
}


func (m mainModel) RenderSelected() string {

	if m.Selected < 0 {
		return m.Render()
	}

	cellWidth := m.WindowSize.Width - 10
	cellHeight := m.WindowSize.Height - 5 

	focusedStyle := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(focusedModelStyle)

	info := m.infos[m.Selected]

	activeStyled := activeStyle.Render("Active")
	if info.Active == false {
		activeStyled = disabledStyle.Render("Disabled")
	}

			
	return lipgloss.JoinHorizontal(lipgloss.Top, focusedStyle.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.Title(), info.Description(), activeStyled)))
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return func() tea.Msg { return "" }
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	//var cmds []tea.Cmd
	switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.WindowSize = msg
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
		}

		if m.Selected < 0 {
			return UpdateChoice(msg, m)
		}
	}

	return UpdateSelected(msg, m)
}

func UpdateChoice(msg tea.Msg, m mainModel) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "tab":
			m.state += 1
			if m.state >= len(m.infos) {
				m.state = 0
			}
		case "n":
			m.infos[m.state].Active =	ToggleActive(m.infos[m.state].Active, m.infos[m.state].Id.String())
		case "enter":
			m.Selected = m.state
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func UpdateSelected(msg tea.Msg, m mainModel) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.Selected = -1
		case "tab":
			//m.state += 1
			//if m.state >= len(m.infos) {
			//	m.state = 0
			//}
		case "n":
			m.infos[m.state].Active =	ToggleActive(m.infos[m.state].Active, m.infos[m.state].Id.String())
		case "enter":
			//m.Selected = m.state
			//return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}


func ToggleActive(current bool, id string) bool {

	QueryMsg := "Query.EnableTrigger"
	if current == true {
		QueryMsg = "Query.DisableTrigger"
	}

	var reply *bool
	err := client.Call(QueryMsg, id, &reply)

	if err != nil {
		log.Fatal("query error:", err)
	}

	if *reply {
		return !current
	}

	return current
}

func (m mainModel) View() string {
	var s string
	if m.Selected >= 0 {
		s += m.RenderSelected()
	} else {
		s += m.Render()
	}

	model := m.currentFocusedModel()
	s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model))
	return s
}

func (m mainModel) currentFocusedModel() string {
	return string(m.state)
}



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

	var err error
	client, err = rpc.Receive()

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


		m := newModel(reply)

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
