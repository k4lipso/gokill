package actions

import (
	"encoding/json"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
)

type Remote struct {
	PeerGroupId string `json:"group"`
	Secret      string `json:"secret"`
	TestSecret  string `json:"testSecret"`
	ActionType
}

func (t Remote) DryExecute() {
	t.ActionChan <- remote.Handler.Broadcast(t.PeerGroupId, t.TestSecret)
}

func (t Remote) Execute() {
	t.ActionChan <- remote.Handler.Broadcast(t.PeerGroupId, t.Secret)
}

func (t Remote) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
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
	}
}
