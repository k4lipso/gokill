package actions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/k4lipso/gokill/internal"
)

type ReadFile struct {
	File string `json:"file"`
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

func (c ReadFile) Execute(payload *internal.Payload) {
	b, err := os.ReadFile(c.File)

	if err != nil {
		c.ActionChan <- fmt.Errorf("Error while reading file: %s", err)
		return
	}

	fileContent := string(b)

	messagePayload, err := internal.CreatePayloadMessage(fileContent).ToPayload()

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
			"file": "/home/user/secrets.txt"
		}
	}
	`
}

func (p ReadFile) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{"file", "string", "absolute path to a file that read.", ""},
	}
}
