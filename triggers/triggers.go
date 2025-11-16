package triggers

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
)

type TriggerState int8

const (
	Initialized TriggerState = iota
	Test
	Armed
	Firing
	Failed
	Done
	Disabled
)

func (e TriggerState) String() string {
	switch e {
	case Initialized:
		return "Initialized"
	case Test:
		return "Test"
	case Armed:
		return "Armed"
	case Firing:
		return "Firing"
	case Failed:
		return "Failed"
	case Done:
		return "Done"
	case Disabled:
		return "Disabled"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

type TriggerDisabledError struct{}

func (e *TriggerDisabledError) Error() string {
	return "Trigger was disabled"
}

type TriggerUpdate struct {
	State   TriggerState
	Trigger Trigger
	Error   error
}

type TriggerUpdateChan chan TriggerUpdate

type Observable struct {
	NotificationChan []TriggerUpdateChan
}

func (o *Observable) Attach(Chan TriggerUpdateChan) {
	o.NotificationChan = append(o.NotificationChan, Chan)
}

func (o Observable) Notify(state TriggerState, trigger Trigger, Error error) {
	o.NotifyUpdate(TriggerUpdate{state, trigger, Error})
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

type TriggerBase struct {
	action  actions.Action
	enabled bool
}

func (t *TriggerBase) Fire() {
	actions.Fire(t.action)
}

func (t *TriggerBase) GetAction() actions.Action {
	return t.action
}

func (t *TriggerBase) IsEnabled() bool {
	return t.enabled
}

func (t *TriggerBase) Enable(state bool) {
	t.enabled = state
}

type Trigger interface {
	internal.Documenter
	Listen() error
	Fire()
	GetAction() actions.Action
	Enable(state bool)
	IsEnabled() bool
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
}

func NewTriggerHandler(config internal.KillSwitchConfig) *TriggerHandler {
	return &TriggerHandler{
		Observable: createObservable(),
		Name:       config.Name,
		Loop:       config.Loop,
		Config:     config,
		Id:         uuid.New(),
		Running:    false,
		State:      Initialized,
	}
}

func (t TriggerHandler) GetName() string {
	return t.WrappedTrigger.GetName()
}

func (t TriggerHandler) GetDescription() string {
	return t.WrappedTrigger.GetDescription()
}

func (t TriggerHandler) GetExample() string {
	return t.WrappedTrigger.GetExample()
}

func (t TriggerHandler) GetOptions() []internal.ConfigOption {
	return t.WrappedTrigger.GetOptions()
}

func (t *TriggerHandler) Create(config internal.KillSwitchConfig) (*TriggerHandler, error) {
	return NewTriggerHandler(config), nil
}

func (t *TriggerHandler) UpdateState(state TriggerState, err error) {
	t.State = state
	t.Notify(state, t.WrappedTrigger, err)
}

func (t *TriggerHandler) Listen() {
	for {
		t.Running = true

		defer func() { t.Running = false }()

		t.UpdateState(Armed, nil)
		t.TimeStarted = time.Now()
		err := t.WrappedTrigger.Listen()

		if errors.Is(err, &TriggerDisabledError{}) {
			t.UpdateState(Disabled, err)
			return
		}

		if err != nil {
			t.UpdateState(Failed, err)
			continue
		}

		t.UpdateState(Firing, nil)
		t.TimeFired = time.Now()
		t.WrappedTrigger.Fire()
		t.UpdateState(Done, nil)

		if !t.Loop {
			return
		}
	}
}

func NewTrigger(config internal.KillSwitchConfig) (*TriggerHandler, error) {
	result := NewTriggerHandler(config)

	for _, availableTrigger := range GetAllTriggers() {
		if config.Type == availableTrigger.GetName() {
			t, err := availableTrigger.Create(config)
			t.Enable(true)

			if err != nil {
				return nil, fmt.Errorf("Could not create Trigger, reason: %s", err)
			}

			result.WrappedTrigger = t
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
