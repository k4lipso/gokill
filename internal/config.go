package internal

import "fmt"

type OptionMissingError struct {
	OptionName string
}

func (o OptionMissingError) Error() string {
	return fmt.Sprintf("Error during config parsing: option %s could not be parsed.", o.OptionName)
}

type Options map[string]interface{}

type ActionConfig struct {
	Type    string  `json:"type"`
	Options Options `json:"options"`
	Stage   int     `json:"stage"`
}

type KillSwitchConfig struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Options Options        `json:"options"`
	Actions []ActionConfig `json:"actions"`
}
