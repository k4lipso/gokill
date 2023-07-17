package triggers

import (
	"fmt"
	"time"

	"unknown.com/gokill/actions"
)

type Options map[string]interface{}

type KillSwitchConfig struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Options Options                `json:"options"`
	Actions []actions.ActionConfig `json:"actions"`
}

type Trigger interface {
	Listen()
}

type TimeOut struct {
	d      time.Duration
	action actions.Action
}

func (t TimeOut) Listen() {
	fmt.Println("TimeOut listens")
	time.Sleep(t.d)
	fmt.Println("TimeOut fires")
	t.action.Execute()
}

type OptionMissingError struct {
	optionName string
}

func (o OptionMissingError) Error() string {
	return fmt.Sprintf("Error during config parsing: option %s could not be parsed.", o.optionName)
}

// func NewTimeOut(d time.Duration, action actions.Action) TimeOut {
func NewTimeOut(config KillSwitchConfig) (TimeOut, error) {
	opts := config.Options

	duration, ok := opts["duration"]

	if !ok {
		return TimeOut{}, OptionMissingError{"duration"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return TimeOut{}, err
	}

	return TimeOut{time.Duration(duration.(float64)) * time.Second, action}, nil
}

func NewTrigger(config KillSwitchConfig) (Trigger, error) {
	if config.Type == "TimeOut" {
		return NewTimeOut(config)
	}

	return nil, fmt.Errorf("Error parsing config: Trigger with type %s does not exists", config.Type)
}
