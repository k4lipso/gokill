package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type OptionMissingError struct {
	OptionName string
}

func (o OptionMissingError) Error() string {
	return fmt.Sprintf("Error during config parsing: option %s could not be parsed.", o.OptionName)
}

type ActionConfig struct {
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
	Stage   int             `json:"stage"`
}

type KillSwitchConfig struct {
	Name    string          `json:"name"`
	Loop    bool            `json:"loop"`
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
	Actions []ActionConfig  `json:"actions"`
}

type ConfigOption struct {
	Name        string
	Type        string
	Description string
	Default     string
}

type Documenter interface {
	GetName() string
	GetDescription() string
	GetExample() string
	GetOptions() []ConfigOption
}

func EnsureDirExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		Log.Infof("Directory created: %s", dir)
	} else if err != nil {
		return fmt.Errorf("error checking directory: %v", err)
	}

	return nil
}
