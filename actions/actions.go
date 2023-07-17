package actions

import (
	"fmt"
	"sort"
	"time"
)

type OptionMissingError struct {
	optionName string
}

func (o OptionMissingError) Error() string {
	return fmt.Sprintf("Error during config parsing: option %s could not be parsed.", o.optionName)
}

type Options map[string]interface{}

type ActionConfig struct {
	Type    string  `json:"type"`
	Options Options `json:"options"`
	Stage   int     `json:"stage"`
}

type KillSwitchConfig struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Options Options        `json:"options"`
	Actions []ActionConfig `json:"actions"`
}

type Action interface {
	Execute()
}

type Printer struct {
	Message    string
	ActionChan chan bool
}

func (p Printer) Execute() {
	fmt.Printf("Print action fires. Message: %s", p.Message)
	p.ActionChan <- true
}

type TimeOut struct {
	Duration   time.Duration
	ActionChan chan bool
}

func (t TimeOut) Execute() {
	fmt.Printf("Waiting %d seconds\n", t.Duration/time.Second)
	time.Sleep(t.Duration)
	t.ActionChan <- true
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

func NewPrint(config ActionConfig, c chan bool) (Action, error) {
	opts := config.Options
	message, ok := opts["message"]

	if !ok {
		return nil, OptionMissingError{"message"}
	}

	return Printer{fmt.Sprintf("%v", message), c}, nil
}

func NewTimeOut(config ActionConfig, c chan bool) (Action, error) {
	opts := config.Options
	duration, ok := opts["duration"]

	if !ok {
		return nil, OptionMissingError{"message"}
	}

	return TimeOut{time.Duration(duration.(float64)) * time.Second, c}, nil
}

func NewSingleAction(config ActionConfig, c chan bool) (Action, error) {
	if config.Type == "Print" {
		return NewPrint(config, c)
	}

	if config.Type == "TimeOut" {
		return NewTimeOut(config, c)
	}

	return nil, fmt.Errorf("Error parsing config: Action with type %s does not exists", config.Type)
}

func NewAction(config []ActionConfig) (Action, error) {
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
