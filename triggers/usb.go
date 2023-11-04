package triggers

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
)

type UsbDisconnect struct {
	WaitTillConnected bool   `json:"waitTillConnected"`
	DeviceName        string `json:"deviceName"`
	action            actions.Action
}

func isUsbConnected(deviceName string) bool {
	devicePath := "/dev/disk/by-id/" + deviceName

	_, err := os.Open(devicePath)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (t UsbDisconnect) Listen() {
	if t.WaitTillConnected {
		for !isUsbConnected(t.DeviceName) {
			time.Sleep(1 * time.Second)
		}

		internal.LogDoc(t).Infof("Device %s detected.", t.DeviceName)
		internal.LogDoc(t).Notice("Trigger is Armed")
	}

	for {
		if !isUsbConnected(t.DeviceName) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	actions.Fire(t.action)
}

func CreateUsbDisconnect(config internal.KillSwitchConfig) (UsbDisconnect, error) {
	result := UsbDisconnect{
		WaitTillConnected: true,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return UsbDisconnect{}, err
	}

	if result.DeviceName == "" {
		return UsbDisconnect{}, internal.OptionMissingError{"deviceName"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return UsbDisconnect{}, err
	}

	result.action = action

	return result, nil
}

func (e UsbDisconnect) Create(config internal.KillSwitchConfig) (Trigger, error) {
	return CreateUsbDisconnect(config)
}

func (p UsbDisconnect) GetName() string {
	return "UsbDisconnect"
}

func (p UsbDisconnect) GetDescription() string {
	return `
Triggers when given usb drive is disconnected.
Currently it simply checks that the file /dev/disk/by-id/$deviceId exists.
`
}

func (p UsbDisconnect) GetExample() string {
	return `
	{
		"type": "UsbDisconnect",
		"name": "Example Trigger",
		"options": {
			"deviceId": "ata-Samsung_SSD_860_EVO_1TB_S4AALKWJDI102",
			"waitTillConnected": true
		},
		"actions": [
		]
	}
	`
}


func (p UsbDisconnect) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"waitTillConnected", "bool", "Only trigger when device was connected before", "true"},
		{"deviceId", "string", "Name of device under /dev/disk/by-id/", "\"\""},
	}
}
