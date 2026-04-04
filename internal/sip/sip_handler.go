package sip

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/emiago/diago"
	"github.com/emiago/diago/media"
	"github.com/emiago/sipgo"
	diaSip "github.com/emiago/sipgo/sip"
	"github.com/k4lipso/gokill/internal"
)

var (
	Handler *SipHandler
)

type SipHandler struct {
	internal.ExternalTriggerMap
	Username      string
	Password      string
	Registrar     string
	RecipientUri  string
	AudioFilePath string
	ListenAddress string
}

type SipConfig struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Registrar     string `json:"registrar"`
	SipUri        string `json:"sipUri"`
	AudioFilePath string `json:"audioFilePath"`
	ListenAddress string `json:"listenAddress"`
}

func CreateSipHandler(ctx context.Context, sipConfigPath string) (SipHandler, error) {
	internal.Log.Info("Start creation of gokill Peerhandler")

	if sipConfigPath == "" {
		return SipHandler{}, fmt.Errorf("Empty sip config path given.")
	}

	configFile, err := os.ReadFile(sipConfigPath)

	if err != nil {
		return SipHandler{}, fmt.Errorf("Could not open sip config file. Reason: %s", err)
	}

	var config SipConfig
	err = json.Unmarshal(configFile, &config)

	if config.Username == "" {
		return SipHandler{}, fmt.Errorf("Empty Username.")
	}

	if config.Password == "" {
		return SipHandler{}, fmt.Errorf("Empty Password.")
	}

	if config.Registrar == "" {
		return SipHandler{}, fmt.Errorf("Empty Registrar.")
	}

	if config.SipUri == "" {
		return SipHandler{}, fmt.Errorf("Empty SipUri.")
	}

	if config.AudioFilePath == "" {
		return SipHandler{}, fmt.Errorf("Empty AudioFilePath.")
	}

	if config.ListenAddress == "" {
		config.ListenAddress = "0.0.0.0"
	}

	result := SipHandler{
		ExternalTriggerMap: internal.ExternalTriggerMap{
			TriggerChannels: make(map[string]internal.TriggerChannel),
		},
		Username:      config.Username,
		Password:      config.Password,
		Registrar:     config.Registrar,
		RecipientUri:  config.SipUri,
		AudioFilePath: config.AudioFilePath,
		ListenAddress: config.ListenAddress,
	}

	return result, nil
}

func (s *SipHandler) Run(ctx context.Context) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		lvl = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(lvl)
	media.RTPDebug = false
	media.RTCPDebug = false
	diaSip.SIPDebug = false
	diaSip.TransactionFSMDebug = false
	recipientUri := s.RecipientUri
	if recipientUri == "" {
		flag.Usage()
		return
	}

	err := s.start(ctx, recipientUri, diago.RegisterOptions{
		Username:  s.Username,
		Password:  s.Password,
		ProxyHost: s.Registrar,
	})

	if err != nil {
		internal.Log.Errorf("PBX finished with error: %s", err)
	}
}

func (s *SipHandler) start(ctx context.Context, recipientURI string, regOpts diago.RegisterOptions) error {
	recipient := diaSip.Uri{}
	if err := diaSip.ParseUri(recipientURI, &recipient); err != nil {
		return fmt.Errorf("failed to parse register uri: %w", err)
	}

	useragent := regOpts.Username
	if useragent == "" {
		useragent = "change-me"
	}

	ua, _ := sipgo.NewUA(
		sipgo.WithUserAgent(useragent),
		sipgo.WithUserAgentHostname(s.ListenAddress),
	)
	defer ua.Close()

	tu := diago.NewDiago(ua, diago.WithTransport(
		diago.Transport{
			Transport: "udp4",
			BindHost:  s.ListenAddress,
			BindPort:  15060,
		},
	))

	go func() {
		tu.Serve(ctx, func(inDialog *diago.DialogServerSession) {
			internal.Log.Infof("New dialog request with id: %s", inDialog.ID)
			defer internal.Log.Infof("Dialog finished. Id: %s", inDialog.ID)
			if err := s.Playback(inDialog); err != nil {
				internal.Log.Errorf("Failed to play: %s", err)
			}
		})
	}()

	return tu.Register(ctx, recipient, regOpts)
}

func (s *SipHandler) Playback(inDialog *diago.DialogServerSession) error {
	inDialog.Trying()
	inDialog.Ringing()
	time.Sleep(time.Second * 3)
	inDialog.Answer()

	playfile, err := os.Open(s.AudioFilePath)
	if err != nil {
		return err
	}

	internal.Log.Infof("Playing a file %s", s.AudioFilePath)

	pb, err := inDialog.PlaybackCreate()
	if err != nil {
		return err
	}
	go func() {
		_, err = pb.Play(playfile, "audio/wav")
		if err != nil {
			internal.Log.Error(err.Error())
			return
		}
	}()

	reader := inDialog.AudioReaderDTMF()

	var dtmfPin string
	return reader.Listen(func(dtmf rune) error {
		internal.Log.Infof("Received DTMF: %s", string(dtmf))
		dtmfPin += string(dtmf)

		event := internal.TriggerEvent{
			Secret:  dtmfPin,
			Payload: nil,
		}

		err := s.ExecuteRemoteTrigger(event)

		if err == nil {
			dtmfPin = ""
		}

		return nil
	}, 10*time.Second)
}
