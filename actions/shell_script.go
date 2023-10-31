package actions

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"os"

	"unknown.com/gokill/internal"
)

type ShellScript struct {
	Path string `json:"path"`
	ActionChan ActionResultChan
}

func isExecutableFile(path string) bool {
	fi, err := os.Lstat(path)

	if err != nil {
		fmt.Println("Test executing Shellscript Failed.")
		return false
	}

	mode := fi.Mode()

	//TODO: should check if current user can execute 
	if mode&01111 == 0 {
		return false
	}

	return true
}

func (c ShellScript) DryExecute() {
	fmt.Printf("Test Executing ShellScript:\n%s\n", c.Path)

	_, err := os.Open(c.Path)

	if err != nil {
		fmt.Println("Test executing Shellscript Failed.")
		c.ActionChan <- err
		return
	}

	if !isExecutableFile(c.Path) {
		fmt.Println("Test executing Shellscript Failed.")
		c.ActionChan <- fmt.Errorf("File is not executable: %s", c.Path)
		return
	}

	c.ActionChan <- nil
}

func (c ShellScript) Execute() {
	if !isExecutableFile(c.Path) {
		fmt.Println("Test executing Shellscript Failed.")
		c.ActionChan <- fmt.Errorf("File is not executable: %s", c.Path)
		return
	}

	cmd := exec.Command("/bin/sh", c.Path)

	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		c.ActionChan <- err
	}

	fmt.Println(string(stdout[:]))
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
		}
	}
	`
}

func (p ShellScript) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"path", "string", "path to script to execute", ""},
	}
}