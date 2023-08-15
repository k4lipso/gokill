package triggers

import (
	"encoding/json"
	"fmt"
	"time"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
)

type TimeOut struct {
	Duration int
	action   actions.Action
}

func (t TimeOut) Listen() {
	fmt.Println("TimeOut listens")
	fmt.Println(t.Duration)
	time.Sleep(time.Duration(t.Duration) * time.Second)
	fmt.Println("TimeOut fires")
	actions.Fire(t.action)
}

func NewTimeOut(config internal.KillSwitchConfig) (TimeOut, error) {
	var result TimeOut
	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return TimeOut{}, err
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return TimeOut{}, err
	}

	result.action = action
	return result, nil
}

func (p TimeOut) GetName() string {
	return "TimeOut"
}

func (p TimeOut) GetDescription() string {
	return "Triggers after given duration."
}

func (p TimeOut) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"duration", "string", "duration in seconds", "0"},
	}
}
