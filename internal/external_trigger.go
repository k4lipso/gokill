package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

type PayloadType string

const (
	PayloadTypeMessage PayloadType = "message"
)

type Payload struct {
	Type PayloadType
	Data any
}

func (p Payload) AsMessage() (PayloadMessage, error) {
	if p.Type != PayloadTypeMessage {
		return PayloadMessage{}, fmt.Errorf("PayloadType is not message. Got: %s", p.Type)
	}

	var result PayloadMessage
	data, ok := p.Data.([]byte)

	if !ok {
		return PayloadMessage{}, fmt.Errorf("Could not convert Data to byte array")
	}

	err := json.Unmarshal(data, &result)

	if err != nil {
		return PayloadMessage{}, fmt.Errorf("Could not decode message from json. Reason: %s", err)
	}

	return result, nil
}

type PayloadMessage struct {
	CreatedAt string `json:"createdAt"`
	Message   string `json:"message"`
}

func CreatePayloadMessage(msg string) PayloadMessage {
	return PayloadMessage{
		CreatedAt: time.Now().String(),
		Message:   msg,
	}
}

func (p PayloadMessage) ToPayload() (Payload, error) {
	data, err := json.Marshal(p)

	return Payload{
		Type: PayloadTypeMessage,
		Data: data,
	}, err
}

type TriggerEvent struct {
	Secret  string   `json:"secret"`
	Payload *Payload `json:"payload,omitempty"`
}

type TriggerChannelEvent struct {
	IsTest bool
	Event  TriggerEvent
}

type TriggerChannel struct {
	IsTest  bool
	Channel chan TriggerChannelEvent
}

type ExternalTrigger interface {
	RegisterRemoteTrigger(secret string, testSecret string) (chan TriggerChannelEvent, error)
}

type ExternalTriggerMap struct {
	TriggerChannels map[string]TriggerChannel
}

func (n *ExternalTriggerMap) RegisterRemoteTrigger(secret string, testSecret string) (chan TriggerChannelEvent, error) {
	if secret == "" || testSecret == "" {
		return nil, fmt.Errorf("Empty secret or testSecret. That is not allowed!")
	}

	triggerChannel, exists := n.TriggerChannels[secret]

	if exists {
		return triggerChannel.Channel, nil
	}

	channel := make(chan TriggerChannelEvent)

	n.TriggerChannels[secret] = TriggerChannel{
		IsTest:  false,
		Channel: channel,
	}

	n.TriggerChannels[testSecret] = TriggerChannel{
		IsTest:  true,
		Channel: channel,
	}

	return channel, nil
}

func (n *ExternalTriggerMap) ExecuteRemoteTrigger(event TriggerEvent) error {
	val, ok := n.TriggerChannels[event.Secret]

	if !ok {
		return fmt.Errorf("Cant execute remote trigger! Trigger with secret: '%s' does not exists.", event.Secret)
	}

	val.Channel <- TriggerChannelEvent{
		IsTest: val.IsTest,
		Event:  event,
	}

	return nil
}
