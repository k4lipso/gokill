package triggers

import (
	"context"
	"testing"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

type MockTrigger struct {
	OnInit   func(context.Context) error
	OnListen func(context.Context) (TriggerState, error)
}

func (m *MockTrigger) Init(ctx context.Context) error {
	return m.OnInit(ctx)
}

func (m *MockTrigger) Listen(ctx context.Context) (TriggerState, error) {
	return m.OnListen(ctx)
}

func (m *MockTrigger) Create(config internal.KillSwitchConfig) (Trigger, error) {
	return &MockTrigger{}, nil
}

func (t *MockTrigger) GetName() string {
	return ""
}

func (t *MockTrigger) GetDescription() string {
	return ""
}

func (t *MockTrigger) GetExample() string {
	return ""
}

func (t *MockTrigger) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{}
}

type MockAction struct {
	gotExecuted bool
	gotTested   bool
}

func (m *MockAction) Execute() {
	m.gotExecuted = true
}

func (m *MockAction) DryExecute() {
	m.gotTested = true
}

func (m *MockAction) GetActionChan() actions.ActionResultChan {
	return nil
}

func (m *MockAction) Create(internal.ActionConfig, actions.ActionResultChan) (actions.Action, error) {
	return &MockAction{}, nil
}

func TestTriggerHandler(t *testing.T) {
	type MockTriggerTest struct {
		trigger                MockTrigger
		action                 MockAction
		loop                   bool
		testRun                bool
		cancelTimeout          int
		expectedFinalState     TriggerState
		expectedError          error
		expectedActionExecuted bool
		expectedActionTested   bool
	}

	triggerList := []MockTriggerTest{
		{
			trigger: MockTrigger{
				OnInit: func(ctx context.Context) error {
					return nil
				},
				OnListen: func(ctx context.Context) (TriggerState, error) {
					return Failed, &TriggerCancelledError{}
				},
			},
			action:                 MockAction{},
			loop:                   false,
			testRun:                true,
			cancelTimeout:          10,
			expectedFinalState:     Failed,
			expectedError:          &TriggerCancelledError{},
			expectedActionExecuted: false,
			expectedActionTested:   false,
		},
		{
			trigger: MockTrigger{
				OnInit: func(ctx context.Context) error {
					return &TriggerCancelledError{}
				},
				OnListen: func(ctx context.Context) (TriggerState, error) {
					return Triggered, nil
				},
			},
			action:                 MockAction{},
			loop:                   false,
			testRun:                true,
			cancelTimeout:          10,
			expectedFinalState:     Failed,
			expectedError:          &TriggerCancelledError{},
			expectedActionExecuted: false,
			expectedActionTested:   false,
		},
		{
			trigger: MockTrigger{
				OnInit: func(ctx context.Context) error {
					return nil
				},
				OnListen: func(ctx context.Context) (TriggerState, error) {
					return Triggered, nil
				},
			},
			action:                 MockAction{},
			loop:                   false,
			testRun:                true,
			cancelTimeout:          10,
			expectedFinalState:     Done,
			expectedError:          nil,
			expectedActionExecuted: false,
			expectedActionTested:   true,
		},
		{
			trigger: MockTrigger{
				OnInit: func(ctx context.Context) error {
					return nil
				},
				OnListen: func(ctx context.Context) (TriggerState, error) {
					return Test, nil
				},
			},
			action:                 MockAction{},
			loop:                   false,
			testRun:                false,
			cancelTimeout:          10,
			expectedFinalState:     Done,
			expectedError:          nil,
			expectedActionExecuted: false,
			expectedActionTested:   true,
		},
		{
			trigger: MockTrigger{
				OnInit: func(ctx context.Context) error {
					return nil
				},
				OnListen: func(ctx context.Context) (TriggerState, error) {
					return Triggered, nil
				},
			},
			action:                 MockAction{},
			loop:                   false,
			testRun:                false,
			cancelTimeout:          10,
			expectedFinalState:     Done,
			expectedError:          nil,
			expectedActionExecuted: true,
			expectedActionTested:   false,
		},
	}

	internal.InitLogger()
	internal.SetVerbose(true)

	for _, test := range triggerList {
		ctx := context.Background()
		config := internal.KillSwitchConfig{Name: "Test Trigger", Loop: test.loop}
		handler := NewTriggerHandler(config)
		handler.TestRun = test.testRun
		handler.WrappedTrigger = &test.trigger
		handler.Action = &test.action

		cancelCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		triggerUpdateChan := make(TriggerUpdateChan, 100)

		var lastUpdate TriggerUpdate

		handler.Attach(triggerUpdateChan)
		err := handler.Run(cancelCtx)

		close(triggerUpdateChan)
		for triggerUpdate := range triggerUpdateChan {
			lastUpdate = triggerUpdate
		}

		if err != test.expectedError {
			t.Errorf("Incorrect Error returned. Got: %s, wanted: %v", err, test.expectedError)
		}

		if lastUpdate.State != test.expectedFinalState {
			t.Errorf("Incorrect State returned. Got: %s, wanted: %v", lastUpdate.State, test.expectedFinalState)
		}

		if test.action.gotExecuted != test.expectedActionExecuted {
			t.Errorf("Incorrect action gotExecuted. Got: %v, wanted: %v", test.action.gotExecuted, test.expectedActionExecuted)
		}

		if test.action.gotTested != test.expectedActionTested {
			t.Errorf("Incorrect action gotTested. Got: %v, wanted: %v", test.action.gotTested, test.expectedActionTested)
		}

		if handler.Running != false {
			t.Errorf("Incorrect Running State. Got: %v, wanted: %v", handler.Running, false)
		}
	}
}

func TestNewTriggerHandler(t *testing.T) {
	config := internal.KillSwitchConfig{Name: "Test Trigger", Loop: true}
	handler := NewTriggerHandler(config)

	if handler.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, handler.Name)
	}
	if handler.Loop != config.Loop {
		t.Errorf("Expected loop %v, got %v", config.Loop, handler.Loop)
	}
	if handler.State != Created {
		t.Errorf("Expected state %v, got %v", Created, handler.State)
	}
}
