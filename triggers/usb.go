package triggers

import (
	"encoding/json"
	"errors"
	"fmt"
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

		fmt.Sprintln("Device %s detected.", t.DeviceName)
		fmt.Println("UsbDisconnect Trigger is Armed")
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
	return "Triggers when given usb drive is disconnected"
}

func (p UsbDisconnect) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"waitTillConnected", "bool", "Only trigger when device was connected before", "true"},
		{"deviceId", "string", "Name of device under /dev/disk/by-id/", "\"\""},
	}
}
