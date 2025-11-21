package triggers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/k4lipso/gokill/internal"
)

type TimeOut struct {
	Duration int
}

func (t *TimeOut) Init(ctx context.Context) error {
	return nil
}

func (t *TimeOut) Listen(ctx context.Context) (TriggerState, error) {
	internal.LogDoc(t).Info("TimeOut listens")
	internal.LogDoc(t).Infof("%d", t.Duration)

	select {
	case <-time.After(time.Duration(t.Duration) * time.Second):
		internal.LogDoc(t).Notice("TimeOut fires")
		return Triggered, nil
	case <-ctx.Done():
		return Cancelled, nil
	}
}

func (t *TimeOut) Create(config internal.KillSwitchConfig) (Trigger, error) {
	result := &TimeOut{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return result, err
	}

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
