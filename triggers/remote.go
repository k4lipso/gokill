package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
)

type Remote struct {
	PeerGroupId   string `json:"group"`
	Secret        string `json:"secret"`
	TestSecret    string `json:"testSecret"`
	RemoteTrigger internal.ExternalTrigger
}

func (t *Remote) Init(ctx context.Context) error {
	if remote.Handler == nil {
		return fmt.Errorf("Remote Trigger failed to initialize, remote handler is not initialized")
	}

	peerGroup := remote.Handler.GetPeerGroupByName(t.PeerGroupId)

	if peerGroup == nil {
		return fmt.Errorf("Remote Trigger failed to initialize, given group was not found")
	}

	t.RemoteTrigger = peerGroup
	return nil
}

func (t *Remote) Listen(ctx context.Context) (TriggerState, *internal.Payload, error) {
	if t.RemoteTrigger == nil {
		return Failed, nil, fmt.Errorf("Remote Trigger failed to listen, remote handler is not initialized")
	}

	channel, err := t.RemoteTrigger.RegisterRemoteTrigger(t.Secret, t.TestSecret)

	if err != nil {
		return Failed, nil, fmt.Errorf("Could not register remote trigger")
	}

	select {
	case event := <-channel:
		if !event.IsTest {
			return Triggered, event.Event.Payload, nil
		} else {
			return Test, event.Event.Payload, nil
		}
	case <-ctx.Done():
		return Cancelled, nil, &TriggerCancelledError{}
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
