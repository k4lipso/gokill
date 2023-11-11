package actions

import (
	"encoding/json"
	"time"

	"github.com/k4lipso/gokill/internal"
)

type TimeOut struct {
	Duration   time.Duration
	ActionChan ActionResultChan
}

func (t TimeOut) DryExecute() {
	t.Execute()
}

func (t TimeOut) Execute() {
	internal.LogDoc(t).Infof("Waiting %d seconds", t.Duration)
	time.Sleep(time.Duration(t.Duration) * time.Second)
	t.ActionChan <- nil
}

func (t TimeOut) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	var result TimeOut
	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return nil, internal.OptionMissingError{"duration"}
	}

	result.ActionChan = c
	return result, nil
}

func (p TimeOut) GetName() string {
	return "Timeout"
}

func (p TimeOut) GetDescription() string {
	return `
Waits given duration in seconds.
This can be used to wait a certain amount of time before continuing to the next Stage
	`
}

func (p TimeOut) GetExample() string {
	return `
	{
		"type": "Timeout",
		"options": {
			"duration": 5
		}
	}
	`
}

func (p TimeOut) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"duration", "int", "duration in seconds", "0"},
	}
}
