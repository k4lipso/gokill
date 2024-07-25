package actions

import (
	"fmt"
	"sort"

	"github.com/k4lipso/gokill/internal"
)

type ActionResultChan chan error

type Action interface {
	Execute()
	DryExecute()
	GetActionChan() ActionResultChan
	Create(internal.ActionConfig, ActionResultChan) (Action, error)
}

type DocumentedAction interface {
	Action
	internal.Documenter
}

type Stage struct {
	Actions []Action
}

type ActionType struct {
	ActionChan ActionResultChan
}

func (a ActionType) GetActionChan() ActionResultChan {
	return a.ActionChan
}

type StagedActions struct {
	StageCount int
	Stages     []Stage
	ActionType
}

func (a StagedActions) executeInternal(f func(Action)) {
	for idx, stage := range a.Stages {
		if idx < a.StageCount {
			continue
		}

		internal.Log.Infof("Execute Stage %v", idx+1)
		for actionidx, _ := range stage.Actions {
			go f(stage.Actions[actionidx])
		}

		for range stage.Actions {
			err := <-a.ActionChan

			if err != nil {
				internal.Log.Errorf("Error occured on Stage %d: %s", idx+1, err)
			}
		}
	}
}

var TestRun bool

func Fire(a Action) {
	if TestRun {
		a.DryExecute()
		return
	}

	a.Execute()
}

func (a StagedActions) DryExecute() {
	a.executeInternal(func(a Action) { a.DryExecute() })
}

func (a StagedActions) Execute() {
	a.executeInternal(func(a Action) { a.Execute() })
}

func (a StagedActions) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return StagedActions{}, nil
}

func NewSingleAction(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	for _, availableAction := range GetAllActions() {
		if config.Type == availableAction.GetName() {
			return availableAction.Create(config, c)
		}
	}

	return nil, fmt.Errorf("Error parsing config: Action with type %s does not exists", config.Type)
}

func NewAction(config []internal.ActionConfig) (Action, error) {
	sort.Slice(config, func(i, j int) bool {
		return config[i].Stage < config[j].Stage
	})

	stagedActions := StagedActions{
		StageCount: 0,
		Stages: []Stage{},
	}

	stagedActions.ActionChan = make(ActionResultChan)

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
}

func GetAllActions() []DocumentedAction {
	return []DocumentedAction{
		Command{},
		Printer{},
		RemoveFiles{},
		ShellScript{},
		Shutdown{},
		SendMatrix{},
		SendTelegram{},
		TimeOut{},
	}
}

func GetActionByType(actionType string) DocumentedAction {
	actions := GetAllActions()

	for _, action := range actions {
		if action.GetName() == actionType {
			return action
		}
	}

	return nil
}

func GetDocumenters() []internal.Documenter {
	var result []internal.Documenter

	for _, action := range GetAllActions() {
		result = append(result, action)
	}

	return result
}
