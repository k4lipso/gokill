package actions

import (
	"fmt"
	"sort"

	"unknown.com/gokill/internal"
)

type Action interface {
	Execute()
}

type Stage struct {
	Actions []Action
}

type StagedActions struct {
	ActionChan chan bool
	StageCount int
	Stages     []Stage
}

func (a StagedActions) Execute() {
	for idx, stage := range a.Stages {
		if idx < a.StageCount {
			continue
		}

		fmt.Printf("Execute Stage %v\n", idx+1)
		for actionidx, _ := range stage.Actions {
			go stage.Actions[actionidx].Execute()
		}

		for range stage.Actions {
			<-a.ActionChan
		}
	}
}

func NewSingleAction(config internal.ActionConfig, c chan bool) (Action, error) {
	if config.Type == "Print" {
		return NewPrint(config, c)
	}

	if config.Type == "TimeOut" {
		return NewTimeOut(config, c)
	}

	if config.Type == "Shutdown" {
		return NewShutdown(config, c)
	}

	return nil, fmt.Errorf("Error parsing config: Action with type %s does not exists", config.Type)
}

func NewAction(config []internal.ActionConfig) (Action, error) {
	if len(config) == 1 {
		return NewSingleAction(config[0], make(chan bool))
	}

	sort.Slice(config, func(i, j int) bool {
		return config[i].Stage < config[j].Stage
	})

	stagedActions := StagedActions{make(chan bool), 0, []Stage{}}

	stageMap := make(map[int][]Action)

	for _, actionCfg := range config {
		newAction, err := NewSingleAction(actionCfg, stagedActions.ActionChan)

		if err != nil {
			return nil, err
		}

		val, exists := stageMap[actionCfg.Stage]

		if !exists {
			stageMap[actionCfg.Stage] = []Action{newAction}
			continue
		}

		stageMap[actionCfg.Stage] = append(val, newAction)
	}

	for _, value := range stageMap {
		stagedActions.Stages = append(stagedActions.Stages, Stage{value})
	}

	return stagedActions, nil
	//return Action{}, fmt.Errorf("Error parsing config: Action with type %s does not exists", config.Type)
}

func GetDocumenters() []internal.Documenter {
	return []internal.Documenter{
		Printer{},
		TimeOut{},
		Shutdown{},
	}
}
