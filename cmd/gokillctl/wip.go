package main

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	//"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

// sessionState is used to track which model is focused
type sessionState uint

const (
	defaultTime              = time.Minute
	timerView   sessionState = iota
	spinnerView
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
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

type TriggerInfo struct {
	title		string
	desc		string
	active	bool
}

type mainModel struct {
	state   int
	infos   []TriggerInfo
	index   int
	WindowSize tea.WindowSizeMsg
	Selected int
}

func newModel(timeout time.Duration) mainModel {
	m := mainModel{state: 0, Selected: -1,}
	m.infos = []TriggerInfo{
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "fdiadadsasdoo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "fdiadadsasdoo", desc: "bar", active: true},
		TriggerInfo{title: "foo", desc: "bar", active: true},
		TriggerInfo{title: "fdiadadsasdoo", desc: "bar", active: true},
	}

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
		if info.active == false {
			activeStyled = disabledStyle.Render("Disabled")
		}

		if m.state == idx {
			horizontal = append(horizontal, focusedStyle.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.title, info.desc, activeStyled)))
		} else {
			horizontal = append(horizontal, style.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.title, info.desc, activeStyled)))
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
	if info.active == false {
		activeStyled = disabledStyle.Render("Disabled")
	}

			
	return lipgloss.JoinHorizontal(lipgloss.Top, focusedStyle.Render(fmt.Sprintf("Name: %s\nDesc: %s\n%v", info.title, info.desc, activeStyled)))
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
			m.infos[m.state].active = !m.infos[m.state].active
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
			m.infos[m.state].active = !m.infos[m.state].active
		case "enter":
			//m.Selected = m.state
			//return m, nil
		}
	}

	return m, tea.Batch(cmds...)
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

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func main() {
	p := tea.NewProgram(newModel(defaultTime))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
