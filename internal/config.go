package internal

import (
	"encoding/json"
	"fmt"
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
