package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"unknown.com/gokill/internal"
)

type TimeOut struct {
	Duration   time.Duration
	ActionChan ActionResultChan
}

func (t TimeOut) DryExecute() {
	t.Execute()
}

func (t TimeOut) Execute() {
	fmt.Printf("Waiting %d seconds\n", t.Duration)
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
	return "When triggered waits given duration before continuing with next stage"
}

func (p TimeOut) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"duration", "int", "duration in seconds", "0"},
	}
}
