package triggers
import (
	"encoding/json"
	"time"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
)

type TimeOut struct {
	Duration int
	action   actions.Action
}

func (t TimeOut) Listen() {
	internal.LogDoc(t).Info("TimeOut listens")
	internal.LogDoc(t).Infof("%d", t.Duration)
	time.Sleep(time.Duration(t.Duration) * time.Second)
	internal.LogDoc(t).Notice("TimeOut fires")
	actions.Fire(t.action)
}

func (t TimeOut) Create(config internal.KillSwitchConfig) (Trigger, error) {
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
	return "Timeout"
}

func (p TimeOut) GetDescription() string {
	return "Triggers after given duration. Mostly used for debugging."
}

func (p TimeOut) GetExample() string {
	return `
	{
		"type": "Timeout",
		"name": "Example Trigger",
		"options": {
			"duration": 5
		},
		"actions": [
		]
	}
	`
}

func (p TimeOut) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"duration", "int", "duration in seconds", "0"},
	}
}
