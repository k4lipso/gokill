package actions

import (
	"fmt"
	"encoding/json"

	"context"
	"errors"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/crypto/cryptohelper"

	"github.com/k4lipso/gokill/internal"
)

type SendMatrix struct {
	Homeserver string `json:"homeserver"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token string `json:"token"`
	DeviceID string `json:"deviceId"`
	Database string `json:"database"`
	RoomId string `json:"roomId"`
	Message string `json:"message"`
	TestMessage string `json:"testMessage"`
	ActionChan ActionResultChan
}

func (s SendMatrix) sendMessage(message string) error {
	client, err := mautrix.NewClient(s.Homeserver, "", "")
	if err != nil {
		return err
	}

	cryptoHelper, err := cryptohelper.NewCryptoHelper(client, []byte("meow"), s.Database)
	if err != nil {
		return err
	}

	client.DeviceID = id.DeviceID(s.DeviceID)

	if len(s.Token) != 0 {
		client.UserID = id.UserID(s.Username)
		client.AccessToken = s.Password
	} else {
		cryptoHelper.LoginAs = &mautrix.ReqLogin{
			Type:       mautrix.AuthTypePassword,
			Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: s.Username},
			Password:   s.Password,
		}
	}

	err = cryptoHelper.Init(context.Background())
	if err != nil {
		return err
	}

	client.Crypto = cryptoHelper

	internal.LogDoc(s).Info("Matrix Client Now running")
	syncCtx, cancelSync := context.WithCancel(context.Background())
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	go func() {
		err = client.SyncWithContext(syncCtx)
		defer syncStopWait.Done()
		if err != nil && !errors.Is(err, context.Canceled) {
			return
		}
	}()

	time.Sleep(5 * time.Second)
	resp, err := client.SendText(context.Background(), id.RoomID(s.RoomId), message)

	if err != nil {
		return fmt.Errorf("Failed to send event")
	} else {
		internal.LogDoc(s).Info("Matrix Client: Message sent")
		internal.LogDoc(s).Infof("Matrix Client: event_id: %s", resp.EventID.String())
	}

	cancelSync()
	syncStopWait.Wait()
	err = cryptoHelper.Close()
	if err != nil {
		return fmt.Errorf("Error closing database")
	}

	return nil
}

func (s SendMatrix) DryExecute() {
	internal.LogDoc(s).Info("SendMatrix: Trying to send test message")
	err := s.sendMessage(s.TestMessage)	

	if err != nil {
		internal.LogDoc(s).Info("SendMatrix: failed to send test message")
	}

	s.ActionChan <- err
}

func (s SendMatrix) Execute() {
	internal.LogDoc(s).Info("SendMatrix: Trying to send message")
	err := s.sendMessage(s.Message)	

	if err != nil {
		internal.LogDoc(s).Info("SendMatrix: failed to send message")
	}

	s.ActionChan <- err
}

func CreateSendMatrix(config internal.ActionConfig, c ActionResultChan) (SendMatrix, error) {
	result := SendMatrix{}

	err := json.Unmarshal(config.Options, &result)

	if err != nil {
		return SendMatrix{}, err
	}

	if result.Homeserver == "" {
		return SendMatrix{}, internal.OptionMissingError{"homeserver"}
	}

	if result.Username == "" {
		return SendMatrix{}, internal.OptionMissingError{"username"}
	}

	if result.Password == "" || result.Token == "" {
		return SendMatrix{}, internal.OptionMissingError{"password or token"}
	}

	if result.DeviceID == "" {
		return SendMatrix{}, internal.OptionMissingError{"deviceId"}
	}

	if result.Database == "" {
		return SendMatrix{}, internal.OptionMissingError{"database"}
	}

	if result.RoomId == "" {
		return SendMatrix{}, internal.OptionMissingError{"roomId"}
	}

	if result.Message == "" {
		return SendMatrix{}, internal.OptionMissingError{"message"}
	}

	if result.TestMessage == "" {
		return SendMatrix{}, internal.OptionMissingError{"testMessage"}
	}

	result.ActionChan = c
	return result, nil
}

func (s SendMatrix) Create(config internal.ActionConfig, c ActionResultChan) (Action, error) {
	return CreateSendMatrix(config, c)
}

func (p SendMatrix) GetName() string {
	return "SendMatrix"
}

func (p SendMatrix) GetDescription() string {
	return "Sends a message to a given room. The user needs to be part of that room already."
}

func (p SendMatrix) GetExample() string {
	return `
	{
		"type": "SendMatrix",
		"options": {
			"homeserver": "matrix.org",
			"username": "testuser",
			"token": "syt_AHAuYW1pA2Rac2NvamRkamVzZaVc_bXaMfbXzzbIvtXKlealN_2jtbs8",
			"deviceId": "ZDBLQAQLJH",
			"database": "/etc/gokill/matrix.db",
			"roomId": "!Balrthajskensaw:matrix.org",
			"message": "attention, intruders got my device!",
			"testMessage": "this is just a test, no worries"
		}
	}
	`
}

func (p SendMatrix) GetOptions() []internal.ConfigOption {
	return []internal.ConfigOption{
		{
			Name: "homeserver",
			Type: "string",
			Description: "homeserver address.",
			Default: "",
		},
		{
			Name: "username",
			Type: "string",
			Description: "username (localpart, wihout homeserver address)",
			Default: "",
		},
		{
			Name: "password",
			Type: "string",
			Description: "password in clear text",
			Default: "",
		},
		{
			Name: "token",
			Type: "string",
			Description: "access token in clear text",
			Default: "",
		},
		{
			Name: "deviceId",
			Type: "string",
			Description: "A device id. Example: ZDBLQAQLJH",
			Default: "",
		},
		{
			Name: "database",
			Type: "string",
			Description: "path to database file, will be created if not existing. this is necessary to sync keys for encryption.",
			Default: "",
		},
		{
			Name: "roomId",
			Type: "string",
			Description: "",
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
