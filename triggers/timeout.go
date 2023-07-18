package triggers

import (
	"fmt"
	"time"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
)

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

// func NewTimeOut(d time.Duration, action actions.Action) TimeOut {
func NewTimeOut(config internal.KillSwitchConfig) (TimeOut, error) {
	opts := config.Options

	duration, ok := opts["duration"]

	if !ok {
		return TimeOut{}, internal.OptionMissingError{"duration"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return TimeOut{}, err
	}

	return TimeOut{time.Duration(duration.(float64)) * time.Second, action}, nil
}
