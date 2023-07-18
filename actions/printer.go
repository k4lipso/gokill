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

func (p Printer) GetName() string {
	return "Print"
}

func (p Printer) GetDescription() string {
	return "When triggered prints the configured message to stdout"
}

func (p Printer) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"message", "string", "Message that should be printed", "\"\""},
	}
}
