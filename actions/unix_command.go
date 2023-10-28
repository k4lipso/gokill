package actions

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"unknown.com/gokill/internal"
)

type Command struct {
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	ActionChan chan bool
}

func (c Command) DryExecute() {
	fmt.Printf("Test Executing Command:\n%s ", c.Command)
	for _, arg := range c.Args {
		fmt.Printf("%s ", arg)
	}

	fmt.Println("")
	c.ActionChan <- true
}

func (c Command) Execute() {
	cmd := exec.Command(c.Command, c.Args...)

	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(stdout[:]))

	c.ActionChan <- true
}

func CreateCommand(config internal.ActionConfig, c chan bool) (Command, error) {
	result := Command{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return Command{}, err
	}

	if result.Command == "" {
		return Command{}, internal.OptionMissingError{"command"}
	}

	result.ActionChan = c

	return result, nil
}

func (cc Command) Create(config internal.ActionConfig, c chan bool) (Action, error) {
	return CreateCommand(config, c)
}

func (p Command) GetName() string {
	return "Command"
}

func (p Command) GetDescription() string {
	return "When triggered executes given command"
}

func (p Command) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"command", "string", "command to execute", ""},
		{"args", "string[]", "args", ""},
	}
}
