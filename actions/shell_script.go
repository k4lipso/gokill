package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/k4lipso/gokill/internal"
)

type ShellScript struct {
	Path          string `json:"path"`
	Args          string `json:"args"`
	PayloadAsArgs bool   `json:"payloadAsArgs"`
	ActionType
}

func isExecutableFile(path string) bool {
	fi, err := os.Lstat(path)

	if err != nil {
		return false
	}

	mode := fi.Mode()

	//TODO: should check if current user can execute
	if mode&01111 == 0 {
		return false
	}

	return true
}

func (c ShellScript) DryExecute(*internal.Payload) {
	internal.LogDoc(c).Infof("Test Executing ShellScript:\n%s", c.Path)

	_, err := os.Open(c.Path)

	if err != nil {
		internal.LogDoc(c).Warning("Test executing Shellscript Failed.")
		c.ActionChan <- err
		return
	}

	if !isExecutableFile(c.Path) {
		internal.LogDoc(c).Warning("Test executing Shellscript Failed.")
		c.ActionChan <- fmt.Errorf("File is not executable: %s", c.Path)
		return
	}

	c.ActionChan <- nil
}

func (c ShellScript) Execute(payload *internal.Payload) {
	if !isExecutableFile(c.Path) {
		internal.LogDoc(c).Warning("Executing Shellscript Failed.")
		c.ActionChan <- fmt.Errorf("File is not executable: %s", c.Path)
		return
	}

	args := strings.Fields(c.Args)

	if c.PayloadAsArgs {
		if payload == nil {
			c.ActionChan <- fmt.Errorf("PayloadAsArgs is enabled, but no payload given")
			return
		}

		message, err := payload.AsMessage()

		if err != nil {
			internal.LogDoc(c).Errorf("Print action could not access payload. Reason: %s", err)
			c.ActionChan <- err
			return
		}

		args = append(args, strings.Fields(message.Message)...)
	}

	cmd := exec.Command("/bin/sh", append([]string{c.Path}, args...)...)

	stdout, err := cmd.Output()

	if err != nil {
		internal.LogDoc(c).Warning("Failed to execute Shellscript")
		c.ActionChan <- fmt.Errorf("Error during ShellScript execute: %s", err)
		return
	}

	internal.LogDoc(c).Infof("Shellscript output:\n%s", string(stdout[:]))
	c.ActionChan <- nil
}

func CreateShellScript(config internal.ActionConfig, c ActionResultChan) (ShellScript, error) {
	result := ShellScript{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return ShellScript{}, err
	}

	if result.Path == "" {
		return ShellScript{}, internal.OptionMissingError{"path"}
	}

	result.ActionChan = c
	return result, nil
}

func (cc ShellScript) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateShellScript(config, c)
}

func (p ShellScript) GetName() string {
	return "ShellScript"
}

func (p ShellScript) GetDescription() string {
	return "Executes the given shell script."
}

func (p ShellScript) GetExample() string {
	return `
	{
		"type": "ShellScript",
		"options": {
			"path": "/path/to/file.sh"
			"args": "hello world"
		}
	}
	`
}

func (p ShellScript) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"path", "string", "path to script to execute", ""},
		{"args", "string", "arguments passed to the script", ""},
		{"payloadAsArgs", "bool", "pass payload as args", "false"},
	}
}
