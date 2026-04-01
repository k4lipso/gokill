package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/sip"
)

type Sip struct {
	Secret     string `json:"pin"`
	TestSecret string `json:"testPin"`
}

func (t *Sip) Init(ctx context.Context) error {
	return nil
}

func (t *Sip) Listen(ctx context.Context) (TriggerState, error) {
	if sip.Handler == nil {
		return Failed, fmt.Errorf("Sip Trigger failed to listen, Sip handler is not initialized")
	}

	channel, err := sip.Handler.RegisterRemoteTrigger(t.Secret, t.TestSecret)

	if err != nil {
		return Failed, fmt.Errorf("Could not register Sip trigger")
	}

	select {
	case msg := <-channel:
		if msg == internal.TriggerMessageTrigger {
			return Triggered, nil
		} else {
			return Test, nil
		}
	case <-ctx.Done():
		return Cancelled, &TriggerCancelledError{}
	}
}

func (t *Sip) Create(config internal.KillSwitchConfig) (Trigger, error) {
	result := &Sip{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return result, err
	}

	if result.Secret == "" {
		return &Sip{}, internal.OptionMissingError{"pin"}
	}

	if result.TestSecret == "" {
		return &Sip{}, internal.OptionMissingError{"testPin"}
	}

	return result, nil
}

func (p Sip) GetName() string {
	return "Sip"
}

func (p Sip) GetDescription() string {
	return "Triggers after message containing configured secret in given remote group is received."
}

func (p Sip) GetExample() string {
	return `
	{
		"type": "Sip",
		"name": "Example sip trigger",
		"options": {
			"sipHanlde": "defaultSip",
			"pin": "1111",
			"testPin": "0000",
		},
		"actions": [
		]
	}
	`
}

func (p Sip) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{
			Name:        "pin",
			Type:        "string",
			Description: "Pin to be received by SIP Handler that triggers action",
			Default:     "",
		},
		{
			Name:        "testPin",
			Type:        "string",
			Description: "Pin to be received by SIP Handler that triggers test",
			Default:     "",
		},
	}
}
