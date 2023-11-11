package triggers

import (
	"fmt"
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/actions"
)

type ReceiveTelegram struct {
	Token string `json:"token"`
	ChatId int64 `json:"chatId"`
	Message string `json:"message"`
	action actions.Action
}

func (s ReceiveTelegram) Listen() {
	bot, err := tgbotapi.NewBotAPI(s.Token)

	if err != nil {
		return //fmt.Errorf("ReceiveTelegram waitForMessage error: %s", err)
	}

	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	msg := tgbotapi.NewMessage(-746157642, "BOT TEST MESSAGE")
	bot.Send(msg)

	chatId := s.ChatId
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { // If we got a message
			if(update.Message.Chat.ID != chatId) {
				internal.LogDoc(s).Debugf("ReceiveTelegram received wrong ChatId. Got %s, wanted %s", update.Message.Chat.ID, s.ChatId)
				continue	
			}

			if(update.Message.Text != s.Message) {
				internal.LogDoc(s).Debug("ReceiveTelegram received wrong Message")
				continue
			}

			internal.LogDoc(s).Info("ReceiveTelegram received secret message")
			actions.Fire(s.action)
		}
	}
}



func CreateReceiveTelegram(config internal.KillSwitchConfig) (ReceiveTelegram, error) {
	result := ReceiveTelegram{
		ChatId: 0,
	}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return ReceiveTelegram{}, fmt.Errorf("Error during CreateReceiveTelegram: %s", err)
	}

	if result.Token == "" {
		return ReceiveTelegram{}, internal.OptionMissingError{"token"}
	}

	if result.ChatId == 0 {
		return ReceiveTelegram{}, internal.OptionMissingError{"chadId"}
	}

	if result.Message == "" {
		return ReceiveTelegram{}, internal.OptionMissingError{"message"}
	}

	action, err := actions.NewAction(config.Actions)

	if err != nil {
		return ReceiveTelegram{}, fmt.Errorf("Error during CreateReceiveTelegram: %s", err)
	}

	result.action = action

	return result, nil
}

func (e ReceiveTelegram) Create(config internal.KillSwitchConfig) (Trigger, error) {
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
			Name: "token",
			Type: "string",
			Description: "telegram bot token (ask botfather)",
			Default: "",
		},
		{
			Name: "chatId",
			Type: "int",
			Description: "chatId of group or chat you want the message be received from.",
			Default: "",
		},
		{
			Name: "message",
			Type: "string",
			Description: "actual message that, when received, fires the trigger",
			Default: "",
		},
	}
}
