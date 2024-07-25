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
	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

var client *rpcExt.Client

var (
	modelStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#004400"))

	focusedModelStyle = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#00ff00"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00aa00"))
)

type mainModel struct {
	state   int
	infos   []rpc.TriggerInfo
	index   int
	WindowSize tea.WindowSizeMsg
	SelectedTrigger int
	SelectedAction int
}

func newModel(infos []rpc.TriggerInfo) mainModel {
	m := mainModel{state: 0, SelectedTrigger: -1, SelectedAction: -1, infos: infos,}

	return m
}

func (m mainModel) GetActionContent(info rpc.TriggerInfo) string {
	var result string

	result += fmt.Sprintf("Active Actions: %d\n", len(info.Config.Actions))

	var horizontal []string

	cellsPerRow := 1
	//cellWidth := (m.WindowSize.Width / cellsPerRow) - int(0.02 * float32(m.WindowSize.Width))
	cellWidth := (m.WindowSize.Width / 2) - int(0.02 * float32(m.WindowSize.Width))
	cellHeight := m.WindowSize.Height / (int(len(m.infos) / cellsPerRow) + 3)
	cellHeight = 5

	style := lipgloss.NewStyle().
		Width(cellWidth).
		Height(cellHeight).
		Inherit(modelStyle).
		BorderForeground(lipgloss.Color("#440000"))
	focusedStyle := lipgloss.NewStyle().
		Width(cellWidth).
		Height(cellHeight).
		Inherit(focusedModelStyle).
		BorderForeground(lipgloss.Color("#ff0000"))

	for idx, action := range info.Config.Actions {
		actionStr := fmt.Sprintf("\nType: %s\nStage: %d\n", action.Type, action.Stage)
		actionStr += fmt.Sprintf("Selected: %d", m.SelectedAction)
		actionStr += fmt.Sprintf("idx: %d", idx)

		actualAction := actions.GetActionByType(action.Type)
		options := actualAction.GetOptions()

		actionStr += fmt.Sprintf("Options:\n")
		for _, option := range options {
			actionStr += fmt.Sprintf("\t%s: %s\n", option.Name, option.Description)
		}

		if m.SelectedAction == idx {
			horizontal = append(horizontal, focusedStyle.Render(actionStr))
		} else {
			horizontal = append(horizontal, style.Render(actionStr))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, horizontal...)
}


func (m mainModel) Render() string {
	var horizontal []string
	var vertical [][]string

	cellsPerRow := 1
	//cellWidth := (m.WindowSize.Width / cellsPerRow) - int(0.02 * float32(m.WindowSize.Width))
	cellWidth := (m.WindowSize.Width / 2) - int(0.02 * float32(m.WindowSize.Width))
	cellHeight := m.WindowSize.Height / (int(len(m.infos) / cellsPerRow) + 3)
	cellHeight = 5

	style := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(modelStyle)
	focusedStyle := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(focusedModelStyle)
	//styleBig := lipgloss.NewStyle().Width(cellWidth).Height(len(m.infos) * cellHeight + 3).Inherit(modelStyle).Align()
	//focusedStyleBig := lipgloss.NewStyle().Width(cellWidth).Height(len(m.infos) * cellHeight + 3).Inherit(focusedModelStyle)

	count := 0

	GetContent := func(info rpc.TriggerInfo) string {
		activeStyled := activeStyle.Render("Running")
		if info.Active == false {
			activeStyled = disabledStyle.Render("Disabled")
		}

		return fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.Title(), info.Description(), activeStyled)
	}

	for idx, info := range m.infos {
		if m.state == idx && m.SelectedTrigger < 0 {
			horizontal = append(horizontal, focusedStyle.Render(GetContent(info)))
		} else {
			horizontal = append(horizontal, style.Render(GetContent(info)))
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

	verticals := lipgloss.JoinVertical(lipgloss.Top, horizontalRendered...)
	var rightSide string
	if m.SelectedTrigger < 0 {
		rightSide = m.GetActionContent(m.infos[m.state])
	} else {
		rightSide = m.GetActionContent(m.infos[m.state])
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, verticals, rightSide)
}


func (m mainModel) RenderSelectedTrigger() string {

	if m.SelectedTrigger < 0 {
		return m.Render()
	}

	cellWidth := m.WindowSize.Width - 10
	cellHeight := m.WindowSize.Height - 5 

	focusedStyle := lipgloss.NewStyle().Width(cellWidth).Height(cellHeight).Inherit(focusedModelStyle)

	info := m.infos[m.SelectedTrigger]

	activeStyled := activeStyle.Render("Running")
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

		if m.SelectedTrigger < 0 {
			return UpdateChoice(msg, m)
		}
	}

	return UpdateSelectedTrigger(msg, m)
}

func UpdateChoice(msg tea.Msg, m mainModel) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd
	m.SelectedAction = -1

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "shitf+tab", "up", "k":
			m.state -= 1
			if m.state < 0 {
				m.state = len(m.infos) - 1
			}
		case "tab", "down", "j":
			m.state += 1
			if m.state >= len(m.infos) {
				m.state = 0
			}
		case "d":
			m.infos[m.state].Active =	ToggleActive(m.infos[m.state].Active, m.infos[m.state].Id.String())
		case "enter", "l", "right":
			m.SelectedTrigger = m.state
			m.SelectedAction = 0
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func UpdateSelectedTrigger(msg tea.Msg, m mainModel) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "h", "left":
			m.SelectedTrigger = -1
			m.SelectedAction = -1
		case "shitf+tab", "up", "k":
			m.SelectedAction -= 1
			if m.SelectedAction < 0 {
				m.SelectedAction = len(m.infos[m.state].Config.Actions) - 1
			}
		case "tab", "down", "j":
			m.SelectedAction += 1
			if m.SelectedAction >= len(m.infos[m.state].Config.Actions) {
				m.SelectedAction = 0
			}
		case "t":
			go TestAction(m.infos[m.state].Config.Actions[m.SelectedAction])
		case "d":
			m.infos[m.state].Active =	ToggleActive(m.infos[m.state].Active, m.infos[m.state].Id.String())
		case "enter":
			//m.SelectedTrigger = m.state
			//return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func TestAction(conf internal.ActionConfig) {
	QueryMsg := "Query.TestAction"

	var reply *error
	err := client.Call(QueryMsg, conf, &reply)

	if err != nil {
		log.Fatal("query error:", err)
	}

	if *reply == nil {
		return
	}

	return
	
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
	s += m.Render()
	//if m.SelectedTrigger >= 0 {
	//	s += m.RenderSelectedTrigger()
	//} else {
	//	s += m.Render()
	//}

	model := m.currentFocusedModel()
	if m.SelectedTrigger < 0 {
		s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • d: enable/disable %s • q: exit\n", model))
	} else {
		s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • t: test action • q: exit\n"))
	}
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
