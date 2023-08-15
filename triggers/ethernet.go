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
		fmt.Println("Ethernet is disconnected")
		return false
	}

	return true
}

func (t EthernetDisconnect) Listen() {
	fmt.Println("EthernetDisconnect listens")

	if t.WaitTillConnected {
		for !isEthernetConnected("enp0s31f6") {
			time.Sleep(1 * time.Second)
		}
	}

	for {
		if !isEthernetConnected("enp0s31f6") {
			fmt.Println("Ethernet is disconnected")
			break
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("EthernetDisconnect fires")
	t.action.Execute()
}

// func NewTimeOut(d time.Duration, action actions.Action) EthernetDisconnect {
func NewEthernetDisconnect(config internal.KillSwitchConfig) (EthernetDisconnect, error) {
	result := EthernetDisconnect{
		WaitTillConnected: true,
	}

	fmt.Println(string(config.Options))
	err := json.Unmarshal(config.Options, &result)

	fmt.Println(result)

	if err != nil {
		return EthernetDisconnect{}, err
	}

	fmt.Println(result.InterfaceName)

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
