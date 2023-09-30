package triggers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
)

type EthernetDisconnect struct {
	WaitTillConnected bool   `json:"waitTillConnected"`
	InterfaceName     string `json:"interfaceName"`
	action            actions.Action
}

func isEthernetConnected(deviceName string) bool {
	content, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/operstate", deviceName))

	if err != nil {
		fmt.Println(err)
		return false
	}

	if string(content[:4]) == "down" {
		return false
	}

	return true
}

func (t EthernetDisconnect) Listen() {
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

	actions.Fire(t.action)
}

func (e EthernetDisconnect) Create(config internal.KillSwitchConfig) (Trigger, error) {
	result := EthernetDisconnect{
		WaitTillConnected: true,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return EthernetDisconnect{}, err
	}

	if result.InterfaceName == "" {
		return EthernetDisconnect{}, internal.OptionMissingError{"interfaceName"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return EthernetDisconnect{}, err
	}

	result.action = action

	return result, nil
}

func (p EthernetDisconnect) GetName() string {
	return "EthernetDisconnect"
}

func (p EthernetDisconnect) GetDescription() string {
	return "Triggers if Ethernetcable is disconnected."
}

func (p EthernetDisconnect) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"waitTillConnected", "bool", "Only trigger when device was connected before", "true"},
		{"interfaceName", "string", "Name of ethernet adapter", "\"\""},
	}
}
