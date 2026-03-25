package internal

import (
	"fmt"
)

type TriggerMessage int

const (
	TriggerMessageTest TriggerMessage = iota
	TriggerMessageTrigger
)

type TriggerChannel struct {
	IsTest  bool
	Channel chan TriggerMessage
}

type ExternalTriggerMap struct {
	TriggerChannels map[string]TriggerChannel
}

func (n *ExternalTriggerMap) RegisterRemoteTrigger(secret string, testSecret string) (chan TriggerMessage, error) {
	if secret == "" || testSecret == "" {
		return nil, fmt.Errorf("Empty secret or testSecret. That is not allowed!")
	}

	triggerChannel, exists := n.TriggerChannels[secret]

	if exists {
		return triggerChannel.Channel, nil
	}

	channel := make(chan TriggerMessage)

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

func (n *ExternalTriggerMap) ExecuteRemoteTrigger(secret string) error {
	val, ok := n.TriggerChannels[secret]

	if !ok {
		return fmt.Errorf("Cant execute remote trigger! Trigger with secret: '%s' does not exists.", secret)
	}

	if val.IsTest {
		val.Channel <- TriggerMessageTest
	} else {
		val.Channel <- TriggerMessageTrigger
	}

	return nil
}
