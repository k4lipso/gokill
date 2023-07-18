package actions

import (
	"fmt"

	"unknown.com/gokill/internal"
)

type Printer struct {
	Message    string
	ActionChan chan bool
}

func (p Printer) Execute() {
	fmt.Printf("Print action fires. Message: %s", p.Message)
	p.ActionChan <- true
}

func NewPrint(config internal.ActionConfig, c chan bool) (Action, error) {
	opts := config.Options
	message, ok := opts["message"]

	if !ok {
		return nil, internal.OptionMissingError{"message"}
	}

	return Printer{fmt.Sprintf("%v", message), c}, nil
}
