package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
)

type Remote struct {
	PeerGroupId string `json:"group"`
	Secret      string `json:"secret"`
	TestSecret  string `json:"testSecret"`
}

func (t *Remote) Init(ctx context.Context) error {
	return nil
}

func (t *Remote) Listen(ctx context.Context) (TriggerState, error) {
	channel, err := remote.Handler.RegisterRemoteTrigger(t.PeerGroupId, t.Secret, t.TestSecret)

	if err != nil {
		return Failed, fmt.Errorf("Could not register remote trigger")
	}

	select {
	case msg := <-channel:
		if msg == remote.TriggerMessageTrigger {
			return Triggered, nil
		} else {
			return Test, nil
		}
	case <-ctx.Done():
		return Cancelled, &TriggerCancelledError{}
	}
}

func (t *Remote) Create(config internal.KillSwitchConfig) (Trigger, error) {
	result := &Remote{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return result, err
	}

	if result.PeerGroupId == "" {
		return &Remote{}, internal.OptionMissingError{"group"}
	}

	if result.Secret == "" {
		return &Remote{}, internal.OptionMissingError{"secret"}
	}

	if result.TestSecret == "" {
		return &Remote{}, internal.OptionMissingError{"testSecret"}
	}

	return result, nil
}

func (p Remote) GetName() string {
	return "Remote"
}

func (p Remote) GetDescription() string {
	return "Triggers after message containing configured secret in given remote group is received."
}

func (p Remote) GetExample() string {
	return `
	{
		"type": "Remote",
		"name": "Example remote trigger",
		"options": {
			"group": "myGroupName",
			"secret": "SECRET-MESSAGE",
			"testSecret": "SECRET-TESTMESSAGE",
		},
		"actions": [
		]
	}
	`
}

func (p Remote) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"group", "string", "peer group name", "76bf03c7-872b-46fc-baab-d49641798a76"},
		{"secret", "string", "shared secret with trigger", ""},
		{"testSecret", "string", "shared test secret with trigger", ""},
	}
}
