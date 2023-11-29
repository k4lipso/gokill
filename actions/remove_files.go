package actions

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/k4lipso/gokill/internal"
)

type RemoveFiles struct {
	Files				[]string   `json:"files"`
	Directories []string   `json:"directories"`
	ActionChan  ActionResultChan
}

func (c RemoveFiles) getRemoveCommand() string {
	command := "srm"
	isAvailable := isCommandAvailable(command)

	if !isAvailable {
		internal.LogDoc(c).Warningf("Command %s not found, falling back to 'rm'", command)
		command = "rm"
	}

	return command
}

func (c RemoveFiles) DryExecute() {
	internal.LogDoc(c).Infof("Test Execute")

	command := c.getRemoveCommand()
	internal.LogDoc(c).Info("The following commands would have been executed:")

	for _, file := range c.Files {
		internal.LogDoc(c).Noticef("%s -f %s", command, file)		
	}

	for _, dir := range c.Directories {
		internal.LogDoc(c).Noticef("%s -rf %s", command, dir)		
	}

	c.ActionChan <- nil
}

func (c RemoveFiles) Execute() {
	internal.LogDoc(c).Infof("Execute")

	command := c.getRemoveCommand()

	for _, file := range c.Files {
		cmd := exec.Command(command, "-fv", file)

		stdout, err := cmd.Output()

		if err != nil {
			internal.LogDoc(c).Errorf("%s", err.Error())
		}

		internal.LogDoc(c).Infof("Try removing %s", file)
		internal.LogDoc(c).Notice(string(stdout))
	}

	for _, dir := range c.Directories {
		cmd := exec.Command(command, "-rfv", dir)

		stdout, err := cmd.Output()

		if err != nil {
			internal.LogDoc(c).Errorf("%s", err.Error())
		}

		internal.LogDoc(c).Infof("Try removing %s", dir)
		internal.LogDoc(c).Notice(string(stdout))
	}

	c.ActionChan <- nil
}

func CreateRemoveFiles(config internal.ActionConfig, c ActionResultChan) (RemoveFiles, error) {
	result := RemoveFiles{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return RemoveFiles{}, fmt.Errorf("Error parsing RemoveFiles: %s", err)
	}

	result.ActionChan = c

	return result, nil
}

func (cc RemoveFiles) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateRemoveFiles(config, c)
}

func (p RemoveFiles) GetName() string {
	return "RemoveFiles"
}

func (p RemoveFiles) GetDescription() string {
	return `
RemoveFiles deletes the given files and directories.
If available "srm" is used, otherwise RemoveFiles falls back to "rm"
	`
}

func (p RemoveFiles) GetExample() string {
	return `
	{
		"type": "RemoveFiles",
		"options": {
			"files": [
				"/home/user/secrets.txt"
			],
			"directories": [
				"/home/user/.gpg",
				"/home/user/.ssh",
				"/home/user/.thunderbird"
			]
		}
	}
	`
}

func (p RemoveFiles) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"files", "[]string", "list of absolute paths of files that should be deleted.", ""},
		{"directories", "[]string", "list of absolute paths of directories that should be deleted.", ""},
	}
}
