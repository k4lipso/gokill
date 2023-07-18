package actions

import (
	"fmt"
	"time"

	"unknown.com/gokill/internal"
)

type TimeOut struct {
	Duration   time.Duration
	ActionChan chan bool
}

func (t TimeOut) Execute() {
	fmt.Printf("Waiting %d seconds\n", t.Duration/time.Second)
	time.Sleep(t.Duration)
	t.ActionChan <- true
}

func NewTimeOut(config internal.ActionConfig, c chan bool) (Action, error) {
	opts := config.Options
	duration, ok := opts["duration"]

	if !ok {
		return nil, internal.OptionMissingError{"duration"}
	}

	return TimeOut{time.Duration(duration.(float64)) * time.Second, c}, nil
}

func (p TimeOut) GetName() string {
	return "TimeOut"
}

func (p TimeOut) GetDescription() string {
	return "When triggered waits given duration before continuing with next stage"
}

func (p TimeOut) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"duration", "string", "duration in seconds", "0"},
	}
}
