package actions

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/k4lipso/gokill/internal"
)

type Command struct {
	Command    string   `json:"command"`
	ActionChan ActionResultChan
}

func isCommandAvailable(name string) bool {
  cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
  if err := cmd.Run(); err != nil {
		return false
  }

  return true
}

func (c Command) DryExecute() {
	internal.LogDoc(c).Infof("Test Executing Command: %s", c.Command)
	command, _, err := c.splitCommandString()

	if err != nil {
		internal.LogDoc(c).Errorf("Error during argument parsing of command '%s'", c.Command)
		c.ActionChan <- fmt.Errorf("Error on Command action: %s", err)
		return
	}

	isAvailable := isCommandAvailable(command)

	if !isAvailable {
		internal.LogDoc(c).Warningf("Command %s not found", command)
		c.ActionChan <- fmt.Errorf("Command %s not found!", command)
		return
	}

	c.ActionChan <- nil
}

func (c Command) splitCommandString() (string, []string, error) {
	splitted := strings.Fields(c.Command)

	if len(splitted) == 0 {
		return "", nil, fmt.Errorf("Command is empty")
	}

	if len(splitted) == 1 {
		return splitted[0], []string(nil), nil
	}

	return splitted[0], splitted[1:], nil
}

func (c Command) Execute() {
	command, args, err := c.splitCommandString()
	internal.LogDoc(c).Infof("Executing command: ", c.Command)

	if err != nil {
		c.ActionChan <- err
		return
	}

	cmd := exec.Command(command, args...)

	stdout, err := cmd.Output()

	if err != nil {
		internal.LogDoc(c).Errorf("%s", err.Error())
		c.ActionChan <- fmt.Errorf("Executing Command '%s' failed: %s", c.Command, err)
	}

	internal.LogDoc(c).Infof("Command Output:\n%s", string(stdout[:]))

	c.ActionChan <- nil
}

func CreateCommand(config internal.ActionConfig, c ActionResultChan) (Command, error) {
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

func (cc Command) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateCommand(config, c)
}

func (p Command) GetName() string {
	return "Command"
}

func (p Command) GetDescription() string {
	return "Invoces given command using exec."
}

func (p Command) GetExample() string {
	return `
	{
		"type": "Command",
		"options": {
			"command": "srm /path/to/file"
		}
	}
	`
}

func (p Command) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"command", "string", "command to execute", ""},
	}
}
