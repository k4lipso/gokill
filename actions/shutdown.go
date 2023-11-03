package actions

import (
	"os/exec"
	"encoding/json"

	"unknown.com/gokill/internal"
)

type Shutdown struct {
	Timeout string `json:"time"`
	ActionChan ActionResultChan
}

func (s Shutdown) DryExecute() {
	internal.LogDoc(s).Infof("shutdown -h %s", s.Timeout)
	internal.LogDoc(s).Info("Test Shutdown executed...")
	s.ActionChan <- nil
}

func (s Shutdown) Execute() {
	if err := exec.Command("shutdown", "-h", s.Timeout).Run(); err != nil {
		internal.LogDoc(s).Errorf("Failed to initiate shutdown: %s", err)
	}

	internal.LogDoc(s).Notice("Shutdown executed...")

	s.ActionChan <- nil
}

func (s Shutdown) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	result := Shutdown{
		Timeout: "now",
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		internal.LogDoc(s).Warning("Parsing Shutdown options failed.")
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
