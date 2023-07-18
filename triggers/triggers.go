package triggers

import (
	"fmt"

	"unknown.com/gokill/internal"
)

type Trigger interface {
	Listen()
}

func NewTrigger(config internal.KillSwitchConfig) (Trigger, error) {
	if config.Type == "TimeOut" {
		return NewTimeOut(config)
	}

	return nil, fmt.Errorf("Error parsing config: Trigger with type %s does not exists", config.Type)
}

func GetDocumenters() []internal.Documenter {
	return []internal.Documenter{
		TimeOut{},
	}
}
