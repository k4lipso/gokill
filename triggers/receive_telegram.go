package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/k4lipso/gokill/internal"
)

type TelegramBotAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type ReceiveTelegram struct {
	Token       string `json:"token"`
	ChatId      int64  `json:"chatId"`
	Message     string `json:"message"`
	TestMessage string `json:"testMessage"`
	bot         TelegramBotAPI
}

func (s *ReceiveTelegram) Init(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(s.Token)

	if err != nil {
		return err
	}

	bot.Debug = false
	s.bot = bot

	return nil
}

func (s *ReceiveTelegram) Listen(ctx context.Context) (TriggerState, *internal.Payload, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	chatId := s.ChatId
	updates := s.bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return Cancelled, nil, &TriggerCancelledError{}
		case update := <-updates:
			{
				if update.Message != nil { // If we got a message
					if update.Message.Chat.ID != chatId {
						internal.LogDoc(s).Debugf("ReceiveTelegram received wrong ChatId. Got %d, wanted %d",
							update.Message.Chat.ID, s.ChatId)
						continue
					}

					if update.Message.Text == s.Message {
						internal.LogDoc(s).Info("ReceiveTelegram received secret message")
						return Triggered, nil, nil
					}

					if update.Message.Text == s.TestMessage {
						internal.LogDoc(s).Info("ReceiveTelegram received test message")
						return Test, nil, nil
					}

					internal.LogDoc(s).Debug("ReceiveTelegram received wrong Message")
				}
			}
		}
	}
}

func CreateReceiveTelegram(config internal.KillSwitchConfig) (*ReceiveTelegram, error) {
	result := &ReceiveTelegram{
		ChatId: 0,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return &ReceiveTelegram{}, fmt.Errorf("Error during CreateReceiveTelegram: %s", err)
	}

	if result.Token == "" {
		return &ReceiveTelegram{}, internal.OptionMissingError{"token"}
	}

	if result.ChatId == 0 {
		return &ReceiveTelegram{}, internal.OptionMissingError{"chadId"}
	}

	if result.Message == "" {
		return &ReceiveTelegram{}, internal.OptionMissingError{"message"}
	}

	if result.TestMessage == "" {
		return &ReceiveTelegram{}, internal.OptionMissingError{"testMessage"}
	}

	return result, nil
}

func (e *ReceiveTelegram) Create(config internal.KillSwitchConfig) (Trigger, error) {
	return CreateReceiveTelegram(config)
}

func (p ReceiveTelegram) GetName() string {
	return "ReceiveTelegram"
}

func (p ReceiveTelegram) GetDescription() string {
	return "Waits for a specific message in a given chat. Once the message is received, the trigger fires."
}

func (p ReceiveTelegram) GetExample() string {
	return `
	{
		"type": "ReceiveTelegram",
		"options": {
			"token": "5349923487:FFGrETxa0pA29d02Akslw-lkwjdA92KAH2",
			"chatId": -832345892,
			"message": "secretmessagethatfiresthetrigger"
		}
	}
	`
}

func (p ReceiveTelegram) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{
			Name:        "token",
			Type:        "string",
			Description: "telegram bot token (ask botfather)",
			Default:     "",
		},
		{
			Name:        "chatId",
			Type:        "int",
			Description: "chatId of group or chat you want the message be received from.",
			Default:     "",
		},
		{
			Name:        "message",
			Type:        "string",
			Description: "actual message that, when received, fires the trigger",
			Default:     "",
		},
		{
			Name:        "testMessage",
			Type:        "string",
			Description: "message that, when received, triggers test",
			Default:     "",
		},
	}
}
