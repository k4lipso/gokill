package actions

import (
	"fmt"
	"os/exec"

	"unknown.com/gokill/internal"
)

type Shutdown struct {
	ActionChan chan bool
}

func (c Shutdown) Execute() {
	if err := exec.Command("shutdown", "-h", "now").Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
	}

	fmt.Println("Shutdown executed...")

	c.ActionChan <- true
}

func NewShutdown(config internal.ActionConfig, c chan bool) (Action, error) {
	return Shutdown{c}, nil
}

func (p Shutdown) GetName() string {
	return "Shutdown"
}

func (p Shutdown) GetDescription() string {
	return "When triggered shuts down the machine"
}

func (p Shutdown) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{}
}
