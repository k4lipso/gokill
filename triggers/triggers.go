package triggers

import (
	"fmt"

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
)

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
	o.NotifyUpdate(TriggerUpdate{ state, trigger, Error, })
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
	return Observable {
		NotificationChan: []TriggerUpdateChan{},
	}
}

type Observabler interface {
	Attach(TriggerUpdateChan)
	Notify(TriggerState, Trigger, error)
	GetLen() int
}

type Trigger interface {
	internal.Documenter
	Listen() error
	Fire()
	Create(internal.KillSwitchConfig) (Trigger, error)
}

type TriggerHandler struct {
	Observable
	Name string
	Loop bool
	WrappedTrigger Trigger
}

func NewTriggerHandler(config internal.KillSwitchConfig) *TriggerHandler {
	return &TriggerHandler{
		Observable: createObservable(),
		Name: config.Name,
		Loop: config.Loop,
	}
}

func (t TriggerHandler)	GetName() string {
	return t.WrappedTrigger.GetName()
}

func (t TriggerHandler)	GetDescription() string {
	return t.WrappedTrigger.GetName()
}

func (t TriggerHandler)	GetExample() string {
	return t.WrappedTrigger.GetExample()
}

func (t TriggerHandler)	GetOptions() []internal.ConfigOption {
	return t.WrappedTrigger.GetOptions()
}

func (t *TriggerHandler) Create(config internal.KillSwitchConfig) (*TriggerHandler, error) {
	return NewTriggerHandler(config), nil
}

func (t *TriggerHandler) Listen() {
	for {
		t.Notify(Armed, t.WrappedTrigger, nil)
		err := t.WrappedTrigger.Listen()

		if err != nil {
			t.Notify(Failed, t.WrappedTrigger, err)
			continue
		}

		t.Notify(Firing, t.WrappedTrigger, nil)
		t.WrappedTrigger.Fire()
		t.Notify(Done, t.WrappedTrigger, nil)

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
	}
}

func GetDocumenters() []internal.Documenter {
	var result []internal.Documenter

	for _, action := range GetAllTriggers() {
		result = append(result, action)
	}

	return result
}
