package triggers

import (
	"fmt"

	"unknown.com/gokill/internal"
)

type Trigger interface {
	Listen()
	Create(internal.KillSwitchConfig) (Trigger, error)
}

type DocumentedTrigger interface {
	internal.Documenter
	Trigger
}

func NewTrigger(config internal.KillSwitchConfig) (Trigger, error) {
	for _, availableTrigger := range GetAllTriggers() {
		if config.Type == availableTrigger.GetName() {
			return availableTrigger.Create(config)
		}
	}

	return nil, fmt.Errorf("Error parsing config: Trigger with type %s does not exists", config.Type)
}

func GetAllTriggers() []DocumentedTrigger {
	return []DocumentedTrigger{
		TimeOut{},
		EthernetDisconnect{},
	}
}

func GetDocumenters() []internal.Documenter {
	var result []internal.Documenter

	for _, action := range GetAllTriggers() {
		result = append(result, action)
	}

	return result
}
