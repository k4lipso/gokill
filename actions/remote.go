package actions

import (
	"encoding/json"
	"fmt"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
)

type Remote struct {
	PeerGroupId string `json:"group"`
	Secret      string `json:"secret"`
	TestSecret  string `json:"testSecret"`
	Message     string `json:"message"`
	TestMessage string `json:"testMessage"`
	ActionType
}

func (t Remote) executeInternal(msg string, secret string) {
	payload, err := internal.CreatePayloadMessage(msg).ToPayload()

	if err != nil {
		t.ActionChan <- err
		return
	}

	message := internal.TriggerEvent{
		Secret:  secret,
		Payload: &payload,
	}

	messageStr, err := json.Marshal(message)

	if err != nil {
		t.ActionChan <- err
		return
	}

	t.ActionChan <- remote.Handler.Broadcast(t.PeerGroupId, string(messageStr))
}

func (t Remote) DryExecute(*internal.Payload) {
	t.executeInternal(t.TestMessage, t.TestSecret)
}

func (t Remote) Execute(*internal.Payload) {
	t.executeInternal(t.Message, t.Secret)
}

func (t Remote) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	if remote.Handler == nil {
		return Remote{}, fmt.Errorf("Failed to create Remote Action: Remote Handler is not initialized")
	}

	var result Remote
	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return Remote{}, err
	}

	if result.PeerGroupId == "" {
		return Remote{}, internal.OptionMissingError{"group"}
	}

	if result.Secret == "" {
		return Remote{}, internal.OptionMissingError{"secret"}
	}

	if result.TestSecret == "" {
		return Remote{}, internal.OptionMissingError{"testSecret"}
	}

	if result.Message == "" {
		return Remote{}, internal.OptionMissingError{"message"}
	}

	if result.TestMessage == "" {
		return Remote{}, internal.OptionMissingError{"testMessage"}
	}

	result.ActionChan = c
	return result, nil
}

func (p Remote) GetName() string {
	return "Remote"
}

func (p Remote) GetDescription() string {
	return `
When executed it sends the secret to the given PeerGroup.
If any remote trigger within the PeerGroup is configured for the specified secret it will be triggered.
	`
}

func (p Remote) GetExample() string {
	return `
	{
		"type": "Remote",
		"options": {
			"group": "myGroupName",
			"secret": "daljqnxliqhlqdpuiwqdklqfhqlkwh"
		}
	}
	`
}

func (p Remote) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"group", "string", "peer group name", "76bf03c7-872b-46fc-baab-d49641798a76"},
		{"secret", "string", "shared secret with trigger", "SECRET-MESSAGE"},
		{"testSecret", "string", "shared test secret with trigger", "TESTSECRET-MESSAGE"},
		{"message", "string", "message to be delivered to external trigger", "The possibility that Adam Weishaupt killed George Washington and took his place, serving as the first US President for two terms, is now confirmed"},
		{"testMessage", "string", "test message to be delivered to external trigger", "Pink Elephant"},
	}
}
