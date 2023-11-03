package actions

import (
	"fmt"
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"unknown.com/gokill/internal"
)

type SendTelegram struct {
	Token string `json:"token"`
	ChatId string `json:"chatId"`
	Message string `json:"message"`
	TestMessage string `json:"testMessage"`
	ActionChan ActionResultChan
}

func (s SendTelegram) sendMessage(message string) error {
	bot, err := tgbotapi.NewBotAPI(s.Token)
	if err != nil {
		return fmt.Errorf("SendTelegram sendMessage error: %s", err)
	}

	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	msg := tgbotapi.NewMessage(-746157642, message)
	_, err = bot.Send(msg)

	if err != nil {
		return fmt.Errorf("SendTelegram sendMessage error: %s", err)
	}

	return nil
}

func (s SendTelegram) DryExecute() {
	internal.LogDoc(s).Info("SendTelegram: Trying to send test message")
	err := s.sendMessage(s.TestMessage)

	if err != nil {
		internal.LogDoc(s).Info("SendTelegram: failed to send test message")
	}

	s.ActionChan <- err
}

func (s SendTelegram) Execute() {
	internal.LogDoc(s).Info("SendTelegram: Trying to send message")
	err := s.sendMessage(s.Message)	

	if err != nil {
		internal.LogDoc(s).Info("SendTelegram: failed to send message")
	}

	s.ActionChan <- err
}

func CreateSendTelegram(config internal.ActionConfig, c ActionResultChan) (SendTelegram, error) {
	result := SendTelegram{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return SendTelegram{}, err
	}

	if result.Token == "" {
		return SendTelegram{}, internal.OptionMissingError{"token"}
	}

	if result.ChatId == "" {
		return SendTelegram{}, internal.OptionMissingError{"chatId"}
	}

	if result.Message == "" {
		return SendTelegram{}, internal.OptionMissingError{"message"}
	}

	if result.TestMessage == "" {
		return SendTelegram{}, internal.OptionMissingError{"testMessage"}
	}

	result.ActionChan = c
	return result, nil
}

func (s SendTelegram) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateSendTelegram(config, c)
}

func (p SendTelegram) GetName() string {
	return "SendTelegram"
}

func (p SendTelegram) GetDescription() string {
	return "Sends a message to a given room. The user needs to be part of that room already."
}

func (p SendTelegram) GetExample() string {
	return `
	{
		"type": "SendTelegram",
		"options": {
			"token": "5349923487:FFGrETxa0pA29d02Akslw-lkwjdA92KAH2",
			"chatId": "-832345892",
			"message": "attention, intruders got my device!",
			"testMessage": "this is just a test, no worries"
		}
	}
	`
}

func (p SendTelegram) GetOptions() []internal.ConfigOption {
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
			Description: "chatId of group or chat you want the message be sent to.",
			Default: "",
		},
		{
			Name: "message",
			Type: "string",
			Description: "actual message that should be sent",
			Default: "",
		},
		{
			Name: "testMessage",
			Type: "string",
			Description: "message sent during test run",
			Default: "",
		},
	}
}
