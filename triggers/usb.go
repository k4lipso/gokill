package triggers

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/k4lipso/gokill/internal"
)

type UsbDisconnect struct {
	WaitTillConnected bool   `json:"waitTillConnected"`
	DeviceName        string `json:"deviceName"`
}

func isUsbConnected(deviceName string) bool {
	devicePath := "/dev/disk/by-id/" + deviceName

	_, err := os.Open(devicePath)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (t *UsbDisconnect) Init(ctx context.Context) error {
	if !t.WaitTillConnected {
		return nil
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			if isUsbConnected(t.DeviceName) {
				internal.LogDoc(t).Infof("Device %s detected.", t.DeviceName)
				return nil
			}
		case <-ctx.Done():
			return &TriggerCancelledError{}
		}
	}
}

func (t *UsbDisconnect) Listen(ctx context.Context) (TriggerState, *internal.Payload, error) {
	for {
		select {
		case <-time.After(1 * time.Second):
			if !isUsbConnected(t.DeviceName) {
				return Triggered, nil, nil
			}
		case <-ctx.Done():
			return Cancelled, nil, &TriggerCancelledError{}
		}
	}
}

func CreateUsbDisconnect(config internal.KillSwitchConfig) (*UsbDisconnect, error) {
	result := &UsbDisconnect{
		WaitTillConnected: true,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return result, err
	}

	if result.DeviceName == "" {
		return result, internal.OptionMissingError{"deviceName"}
	}

	return result, nil
}

func (e *UsbDisconnect) Create(config internal.KillSwitchConfig) (Trigger, error) {
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
