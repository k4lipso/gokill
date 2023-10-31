package actions

import (
	"fmt"
	"os/exec"
	"encoding/json"

	"unknown.com/gokill/internal"
)

type Shutdown struct {
	Timeout string `json:"time"`
	ActionChan ActionResultChan
}

func (s Shutdown) DryExecute() {
	fmt.Printf("shutdown -h %s\n", s.Timeout)
	fmt.Println("Test Shutdown executed...")
	s.ActionChan <- nil
}

func (s Shutdown) Execute() {
	if err := exec.Command("shutdown", "-h", s.Timeout).Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
	}

	fmt.Println("Shutdown executed...")

	s.ActionChan <- nil
}

func (s Shutdown) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	result := Shutdown{
		Timeout: "now",
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		fmt.Println("Parsing Shutdown options failed.")
		return Shutdown{}, err
	}

	result.ActionChan = c
	return result, nil
}

func (p Shutdown) GetName() string {
	return "Shutdown"
}

func (p Shutdown) GetDescription() string {
	return "Shutsdown the machine by perfoming a ```shutdown -h now```"
}

func (p Shutdown) GetExample() string {
	return `
	{
		"type": "Shutdown",
		"options": {
			"time": "+5" //wait 5 minutes before shutdown
		}
	}
	`
}

func (p Shutdown) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{
			Name: "time",
			Type: "string",
			Description: "TIME parameter passed to shutdown as follows ```shutdown -h TIME```",
			Default: "now",
		},
	}
}
