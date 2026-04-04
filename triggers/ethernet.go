package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/k4lipso/gokill/internal"
)

type EthernetDisconnect struct {
	WaitTillConnected bool   `json:"waitTillConnected"`
	InterfaceName     string `json:"interfaceName"`
}

func isEthernetConnected(deviceName string) bool {
	content, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/operstate", deviceName))

	if err != nil {
		internal.LogDoc(EthernetDisconnect{}).Errorf("Cant read devices operstate. Check the deviceName. error: %s", err)
		return false
	}

	if string(content[:4]) == "down" {
		return false
	}

	return true
}

func (t *EthernetDisconnect) Init(ctx context.Context) error {
	if !t.WaitTillConnected {
		return nil
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			if isEthernetConnected(t.InterfaceName) {
				return nil
			}
		case <-ctx.Done():
			return &TriggerCancelledError{}
		}
	}
}

func (t *EthernetDisconnect) Listen(ctx context.Context) (TriggerState, *internal.Payload, error) {
	for {
		select {
		case <-time.After(1 * time.Second):
			if !isEthernetConnected(t.InterfaceName) {
				return Triggered, nil, nil
			}
		case <-ctx.Done():
			return Cancelled, nil, &TriggerCancelledError{}
		}
	}
}

func CreateEthernetDisconnect(config internal.KillSwitchConfig) (*EthernetDisconnect, error) {
	result := &EthernetDisconnect{
		WaitTillConnected: true,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return &EthernetDisconnect{}, err
	}

	if result.InterfaceName == "" {
		return &EthernetDisconnect{}, internal.OptionMissingError{"interfaceName"}
	}

	return result, nil
}

func (e *EthernetDisconnect) Create(config internal.KillSwitchConfig) (Trigger, error) {
	return CreateEthernetDisconnect(config)
}

func (p EthernetDisconnect) GetName() string {
	return "EthernetDisconnect"
}

func (p EthernetDisconnect) GetDescription() string {
	return "Triggers if Ethernetcable is disconnected."
}

func (p EthernetDisconnect) GetExample() string {
	return `
	{
		"type": "EthernetDisconnect",
		"name": "Example Trigger",
		"options": {
			"interfaceName": "eth0",
			"waitTillConnected": true
		},
		"actions": [
		]
	}
	`
}

func (p EthernetDisconnect) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"waitTillConnected", "bool", "Only trigger when device was connected before", "true"},
		{"interfaceName", "string", "Name of ethernet adapter", "\"\""},
	}
}
