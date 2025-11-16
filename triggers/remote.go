package triggers

import (
	"encoding/json"
	"fmt"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
)

type Remote struct {
	TriggerBase
	PeerGroupId string `json:"group"`
	Secret      string `json:"secret"`
	TestSecret  string `json:"testSecret"`
	RemoteChan  chan bool
}

func (t *Remote) Listen() error {
	channel, err := remote.Handler.RegisterRemoteTrigger(t.PeerGroupId, t.Secret, t.TestSecret)

	if err != nil {
		return fmt.Errorf("Could not register remote trigger")
	}

	//block till message received
	//TODO: it should be evaluated if testSecret or secret.
	<-channel

	if !t.enabled {
		return &TriggerDisabledError{}
	}

	return nil
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

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return result, err
	}

	result.action = action

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
		{"secret", "string", "shared secret with trigger", "SECRET-MESSAGE"},
		{"testSecret", "string", "shared test secret with trigger", "SECRET-TESTMESSAGE"},
	}
}
