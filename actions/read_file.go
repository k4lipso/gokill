package actions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/k4lipso/gokill/internal"
)

type ReadFile struct {
	File        string `json:"file"`
	ReadOnInit  bool   `json:"readOnInit"`
	FileContent string
	ActionType
}

func (c ReadFile) DryExecute(payload *internal.Payload) {
	messagePayload, err := internal.CreatePayloadMessage("ReadFile TEST").ToPayload()

	if err != nil {
		c.ActionChan <- fmt.Errorf("Error while creating message payload: %s", err)
		return
	}

	*payload = messagePayload
	c.ActionChan <- nil
}

func (c *ReadFile) setFileContent() error {
	b, err := os.ReadFile(c.File)

	if err != nil {
		return err
	}

	c.FileContent = string(b)
	return nil
}

func (c ReadFile) Execute(payload *internal.Payload) {
	if !c.ReadOnInit {
		err := c.setFileContent()
		if err != nil {
			c.ActionChan <- fmt.Errorf("Error while reading file: %s", err)
			return
		}
	}

	messagePayload, err := internal.CreatePayloadMessage(c.FileContent).ToPayload()

	if err != nil {
		c.ActionChan <- fmt.Errorf("Error while creating message payload: %s", err)
		return
	}

	*payload = messagePayload
	c.ActionChan <- nil
}

func CreateReadFile(config internal.ActionConfig, c ActionResultChan) (ReadFile, error) {
	result := ReadFile{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return ReadFile{}, fmt.Errorf("Error parsing ReadFile: %s", err)
	}

	if result.File == "" {
		return ReadFile{}, internal.OptionMissingError{"file"}
	}

	if result.ReadOnInit {
		err := result.setFileContent()
		if err != nil {
			return ReadFile{}, fmt.Errorf("Error reading file: %s", err)
		}
	}

	result.ActionChan = c

	return result, nil
}

func (cc ReadFile) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateReadFile(config, c)
}

func (p ReadFile) GetName() string {
	return "ReadFile"
}

func (p ReadFile) GetDescription() string {
	return `
ReadFile reads the given file and attaches its content to a message payload.  
	`
}

func (p ReadFile) GetExample() string {
	return `
	{
		"type": "ReadFile",
		"options": {
			"file": "/home/user/secrets.txt",
			"readOnInit": false
		}
	}
	`
}

func (p ReadFile) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"file", "string", "absolute path to a file that read.", ""},
		{"readOnOinit", "bool", "Read file content during initialization instead during execution", "false"},
	}
}
