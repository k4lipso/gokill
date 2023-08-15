package actions

import (
	"encoding/json"
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

func (p Printer) DryExecute() {
	fmt.Printf("Print action fire test. Message: %s", p.Message)
	p.ActionChan <- true
}

func NewPrint(config internal.ActionConfig, c chan bool) (Action, error) {
	var result Printer
	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return nil, internal.OptionMissingError{"message"}
	}

	result.ActionChan = c
	return result, nil
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
