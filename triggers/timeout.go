package triggers
import (
	"encoding/json"
	"time"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

type TimeOut struct {
	Observable
	Duration int
	action   actions.Action
}

func (t *TimeOut) Listen() {
	t.Notify(Armed, t, nil)
	internal.LogDoc(t).Info("TimeOut listens")
	internal.LogDoc(t).Infof("%d", t.Duration)
	time.Sleep(time.Duration(t.Duration) * time.Second)
	internal.LogDoc(t).Notice("TimeOut fires")

	t.Notify(Firing, t, nil)
	actions.Fire(t.action)
	t.Notify(Done, t, nil)
}

func (t *TimeOut) Create(config internal.KillSwitchConfig) (Trigger, error) {
	result := &TimeOut{
		Observable: createObservable(),
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return result, err
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return result, err
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
