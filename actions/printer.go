package actions

import (
	"encoding/json"
	"fmt"

	"unknown.com/gokill/internal"
)

type Printer struct {
	Message    string
	ActionChan ActionResultChan
}

func (p Printer) Execute() {
	fmt.Printf("Print action fires. Message: %s\n", p.Message)
	p.ActionChan <- nil
}

func (p Printer) DryExecute() {
	fmt.Printf("Print action fire test. Message: %s\n", p.Message)
	p.ActionChan <- nil
}

func (p Printer) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
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
	return `
Prints a given message to stdout.
This action is mostly used for debugging purposes.
	`
}

func (p Printer) GetExample() string {
	return `
		{
			type: "Print",
			"options": {
				"message": "Hello World!"
			}
		}
	`
}

func (p Printer) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{
			Name: "message", 
			Type: "string", 
			Description: "Message that should be printed", 
			Default: "\"\"",
		},
	}
}
