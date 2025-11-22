package triggers

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/k4lipso/gokill/internal"
)

func TestListen(t *testing.T) {
	testListenTriggers := []TriggerCancelTest{
		{
			trigger: &ReceiveTelegram{
				ChatId:      23,
				Message:     "Secret",
				TestMessage: "TestSecret",
				bot: &MockTelegramBotAPI{
					messages: []tgbotapi.Message{
						{
							Chat: &tgbotapi.Chat{
								ID: 23,
							},
							Text: "wrong Message",
						},
					},
				},
			},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
		{
			trigger: &ReceiveTelegram{
				ChatId:      23,
				Message:     "Secret",
				TestMessage: "TestSecret",
				bot: &MockTelegramBotAPI{
					messages: []tgbotapi.Message{
						{
							Chat: &tgbotapi.Chat{
								ID: 0,
							},
							Text: "Secret",
						},
					},
				},
			},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
		{
			trigger: &ReceiveTelegram{
				ChatId:      23,
				Message:     "Secret",
				TestMessage: "TestSecret",
				bot: &MockTelegramBotAPI{
					messages: []tgbotapi.Message{
						{
							Chat: &tgbotapi.Chat{
								ID: 23,
							},
							Text: "Secret",
						},
					},
				},
			},
			expectedError: nil,
			expectedState: Triggered,
		},
		{
			trigger: &ReceiveTelegram{
				ChatId:      23,
				Message:     "Secret",
				TestMessage: "TestSecret",
				bot: &MockTelegramBotAPI{
					messages: []tgbotapi.Message{
						{
							Chat: &tgbotapi.Chat{
								ID: 23,
							},
							Text: "TestSecret",
						},
					},
				},
			},
			expectedError: nil,
			expectedState: Test,
		},
		{
			trigger: &ReceiveTelegram{
				ChatId:      23,
				Message:     "Secret",
				TestMessage: "TestSecret",
				bot:         &MockTelegramBotAPI{},
			},
			expectedError: &TriggerCancelledError{},
			expectedState: Cancelled,
		},
	}

	internal.InitLogger()
	internal.SetVerbose(true)
	ctx := context.Background()
	for _, test := range testListenTriggers {
		cancelCtx, cancel := context.WithCancel(ctx)

		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		state, err := test.trigger.Listen(cancelCtx)

		if err != test.expectedError {
			t.Errorf("Incorrect Error returned. Got: %s, wanted: %v", err, test.expectedError)
		}

		if state != test.expectedState {
			t.Errorf("Incorrect State returned. Got: %s, wanted: %v", state, test.expectedState)
		}
	}
}
