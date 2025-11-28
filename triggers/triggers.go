package triggers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

type TriggerState int8

const (
	Created TriggerState = iota
	Initializing
	Initialized
	Test
	Armed
	Triggered
	Firing
	Failed
	Done
	Disabled
	Cancelled
)

type TriggerEvent struct {
	State TriggerState
	Error error
}

func (e TriggerState) String() string {
	switch e {
	case Created:
		return "Created"
	case Initializing:
		return "Initializing"
	case Initialized:
		return "Initialized"
	case Test:
		return "Test"
	case Armed:
		return "Armed"
	case Triggered:
		return "Triggered"
	case Firing:
		return "Firing"
	case Failed:
		return "Failed"
	case Done:
		return "Done"
	case Disabled:
		return "Disabled"
	case Cancelled:
		return "Cancelled"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

type TriggerCancelledError struct{}

func (e *TriggerCancelledError) Error() string {
	return "Trigger was cancelled."
}

type TriggerDisabledError struct{}

func (e *TriggerDisabledError) Error() string {
	return "Trigger was disabled"
}

type TriggerUpdate struct {
	TriggerEvent
	Trigger Trigger
}

type TriggerUpdateChan chan TriggerUpdate

type Observable struct {
	NotificationChan []TriggerUpdateChan
}

func (o *Observable) Attach(Chan TriggerUpdateChan) {
	o.NotificationChan = append(o.NotificationChan, Chan)
}

func (o Observable) Notify(state TriggerState, trigger Trigger, Error error) {
	o.NotifyUpdate(TriggerUpdate{
		TriggerEvent: TriggerEvent{State: state, Error: Error},
		Trigger:      trigger,
	})
}

func (o Observable) NotifyUpdate(update TriggerUpdate) {
	for _, Chan := range o.NotificationChan {
		Chan <- update
	}
}

func (o Observable) GetLen() int {
	return len(o.NotificationChan)
}

func createObservable() Observable {
	return Observable{
		NotificationChan: []TriggerUpdateChan{},
	}
}

type Observabler interface {
	Attach(TriggerUpdateChan)
	Notify(TriggerState, Trigger, error)
	GetLen() int
}

type Trigger interface {
	//TODO: rm internal.Documenter, make DocumentedTrigger
	internal.Documenter
	Init(context.Context) error
	Listen(context.Context) (TriggerState, error)
	Create(internal.KillSwitchConfig) (Trigger, error)
}

type TriggerHandler struct {
	Observable
	Name           string
	Loop           bool
	WrappedTrigger Trigger
	Config         internal.KillSwitchConfig
	TimeStarted    time.Time
	TimeFired      time.Time
	Id             uuid.UUID
	Running        bool
	State          TriggerState
	Action         actions.Action
	TestRun        bool
}

func NewTriggerHandler(config internal.KillSwitchConfig) *TriggerHandler {
	return &TriggerHandler{
		Observable: createObservable(),
		Name:       config.Name,
		Loop:       config.Loop,
		Config:     config,
		Id:         uuid.New(),
		Running:    false,
		State:      Created,
		TestRun:    false,
	}
}

func (t *TriggerHandler) GetName() string {
	return t.WrappedTrigger.GetName()
}

func (t *TriggerHandler) GetDescription() string {
	return t.WrappedTrigger.GetDescription()
}

func (t *TriggerHandler) GetExample() string {
	return t.WrappedTrigger.GetExample()
}

func (t *TriggerHandler) GetOptions() []internal.ConfigOption {
	return t.WrappedTrigger.GetOptions()
}

func (t *TriggerHandler) Create(config internal.KillSwitchConfig) (*TriggerHandler, error) {
	return NewTriggerHandler(config), nil
}

func (t *TriggerHandler) UpdateState(state TriggerState, err error) {
	t.State = state
	t.Notify(state, t.WrappedTrigger, err)
}

func (t *TriggerHandler) Run(ctx context.Context) error {
	for {
		t.Running = true
		defer func() { t.Running = false }()

		t.UpdateState(Initializing, nil)
		err := t.WrappedTrigger.Init(ctx)

		if err != nil {
			t.UpdateState(Failed, err)
			return err
		}

		t.UpdateState(Initialized, nil)

		ch := make(chan TriggerEvent)

		go func() {
			defer close(ch)
			state, err := t.WrappedTrigger.Listen(ctx)
			ch <- TriggerEvent{State: state, Error: err}
		}()

		t.UpdateState(Armed, nil)
		t.TimeStarted = time.Now()

		var event TriggerEvent
		select {
		case event = <-ch:
			if event.State == Failed {
				t.UpdateState(event.State, event.Error)
				return event.Error
			}
		case <-ctx.Done():
			t.UpdateState(Cancelled, nil)
			return nil
		}

		t.UpdateState(Firing, nil)
		t.TimeFired = time.Now()
		if event.State == Test || t.TestRun {
			t.Action.DryExecute()
		} else if event.State == Triggered {
			t.Action.Execute()
		}

		t.UpdateState(Done, nil)

		if !t.Loop {
			return nil
		}
	}
}

func NewTrigger(config internal.KillSwitchConfig) (*TriggerHandler, error) {
	result := NewTriggerHandler(config)

	for _, availableTrigger := range GetAllTriggers() {
		if config.Type == availableTrigger.GetName() {
			t, err := availableTrigger.Create(config)

			if err != nil {
				return nil, fmt.Errorf("Could not create Trigger, reason: %s", err)
			}

			result.WrappedTrigger = t

			action, err := actions.NewAction(config.Actions)

			if err != nil {
				return result, err
			}

			result.Action = action

			return result, nil
		}
	}

	return nil, fmt.Errorf("Error parsing config: Trigger with type %s does not exists", config.Type)
}

func GetAllTriggers() []Trigger {
	return []Trigger{
		&EthernetDisconnect{},
		&ReceiveTelegram{},
		&TimeOut{},
		&UsbDisconnect{},
		&Remote{},
	}
}

func GetDocumenters() []internal.Documenter {
	var result []internal.Documenter

	for _, trigger := range GetAllTriggers() {
		result = append(result, trigger)
	}

	return result
}
