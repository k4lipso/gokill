package triggers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

type EthernetDisconnect struct {
	Observable
	WaitTillConnected bool   `json:"waitTillConnected"`
	InterfaceName     string `json:"interfaceName"`
	action            actions.Action
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

func (t *EthernetDisconnect) Listen() {
	t.Notify(Armed, t, nil)

	if t.WaitTillConnected {
		for !isEthernetConnected(t.InterfaceName) {
			time.Sleep(1 * time.Second)
		}
	}

	for {
		if !isEthernetConnected(t.InterfaceName) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	t.Notify(Firing, t, nil)
	actions.Fire(t.action)
	t.Notify(Done, t, nil)
}

func CreateEthernetDisconnect(config internal.KillSwitchConfig) (*EthernetDisconnect, error) {
	result := &EthernetDisconnect{
		Observable: createObservable(),
		WaitTillConnected: true,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return &EthernetDisconnect{}, err
	}

	if result.InterfaceName == "" {
		return &EthernetDisconnect{}, internal.OptionMissingError{"interfaceName"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return &EthernetDisconnect{}, err
	}

	result.action = action

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
