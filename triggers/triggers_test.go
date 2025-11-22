package triggers

import (
	"context"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TriggerCancelTest struct {
	trigger       Trigger
	expectedError error
	expectedState TriggerState
}

type MockTelegramBotAPI struct {
	messages []tgbotapi.Message
}

func (m *MockTelegramBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	ch := make(chan tgbotapi.Update, len(m.messages))

	for _, msg := range m.messages {
		ch <- tgbotapi.Update{
			Message: &msg,
		}
	}

	return ch
}

func TestListenTriggerCancellation(t *testing.T) {
	testListenTriggers := []TriggerCancelTest{
		{
			trigger:       &EthernetDisconnect{},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
		{
			trigger:       &UsbDisconnect{},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
		{
			trigger:       &UsbDisconnect{},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
		{
			trigger: &ReceiveTelegram{
				bot: &MockTelegramBotAPI{},
			},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
	}

	ctx := context.Background()
	for _, test := range testListenTriggers {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		state, err := test.trigger.Listen(cancelCtx)

		if err != test.expectedError {
			t.Errorf("Incorrect Error returned. Got: %s, wanted: %s", err, test.expectedError)
		}

		if state != test.expectedState {
			t.Errorf("Incorrect State returned. Got: %s, wanted: %s", state, test.expectedState)
		}
	}
}

func TestInitTriggerCancellation(t *testing.T) {
	testInitTriggers := []TriggerCancelTest{
		{
			trigger: &EthernetDisconnect{
				WaitTillConnected: true,
			},
			expectedError: &TriggerCancelledError{},
		},
		{
			trigger: &EthernetDisconnect{
				WaitTillConnected: false,
			},
			expectedError: nil,
		},
		{
			trigger: &UsbDisconnect{
				WaitTillConnected: true,
			},
			expectedError: &TriggerCancelledError{},
		},
		{
			trigger: &UsbDisconnect{
				WaitTillConnected: false,
			},
			expectedError: nil,
		},
	}

	ctx := context.Background()
	for _, test := range testInitTriggers {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		err := test.trigger.Init(cancelCtx)

		if err != test.expectedError {
			t.Errorf("Incorrect Error returned. Got: %s, wanted: %s", err, test.expectedError)
		}
	}
}
