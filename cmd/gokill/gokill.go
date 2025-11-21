package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/age"
	"github.com/k4lipso/gokill/internal/remote"
	"github.com/k4lipso/gokill/rpc"
	"github.com/k4lipso/gokill/triggers"
)

func GetDocumentation() string {
	actions := actions.GetDocumenters()

	var result string

	writeOptions := func(documenters []internal.Documenter) {
		for _, act := range documenters {
			result += fmt.Sprintf("\n### %v\nDescription: %v  \nValues:\n", act.GetName(), act.GetDescription())

			for _, opt := range act.GetOptions() {
				result += fmt.Sprintf("- Name: **%v**\n\t- Type: %v\n\t- Descr: %v\n\t- Default: %v\n",
					opt.Name, opt.Type, opt.Description, opt.Default)
				result += "\n\n"
			}
		}
	}

	result = "# Available Triggers:\n\n"
	writeOptions(triggers.GetDocumenters())
	result += "\n\n# Available Actions:\n\n"
	writeOptions(actions)

	return result
}

func Observe(c triggers.TriggerUpdateChan) {
	for update := range c {
		switch update.State {
		case triggers.Initialized:
			internal.Log.Debugf("Trigger %s initialized", update.Trigger.GetName())
		case triggers.Armed:
			internal.Log.Debugf("Trigger %s armed", update.Trigger.GetName())
		case triggers.Firing:
			internal.Log.Debugf("Trigger %s firing", update.Trigger.GetName())
		case triggers.Failed:
			internal.Log.Debugf("Trigger %s failed. Reason: %s", update.Trigger.GetName(), update.Error)
		case triggers.Done:
			internal.Log.Debugf("Trigger %s done", update.Trigger.GetName())
		}
	}
}

var (
	dbPath = flag.String("db", "./db", "db file path")
)

func runRemoteHandler(ctx context.Context) {
	internal.Log.Info("Initializing gokill remote handler")
	internal.Log.Info("Looking for Keys...")
	key, err := age.LoadOrGenerateKeys(*dbPath + "/age.key")

	if err != nil {
		internal.Log.Panic(err.Error())
	}

	internal.Log.Infof("Found Key: %s", key.Recipient().String())
	internal.Log.Info("Setting up DHT...")

	h, dht, err := remote.SetupLibp2pHost(ctx, *dbPath)

	if err != nil {
		internal.Log.Panic(err.Error())
	}

	internal.Log.Infof("Own ID: %s", h.ID().String())

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		internal.Log.Panic(err.Error())
	}

	peerHandler := remote.PeerHandler{
		Ctx:    ctx,
		Host:   h,
		PubSub: ps,
		Key:    key,
	}

	configPath := *dbPath + "/config.json"
	internal.Log.Infof("Loading config from: %s", configPath)
	Cfg, err := peerHandler.NewConfig(configPath)

	if err != nil {
		internal.Log.Fatal(err.Error())
	}

	peerHandler.Config = Cfg
	peerHandler.ConfigPath = configPath

	internal.Log.Infof("Setting up PeerGroups...")
	peerHandler.InitPeerGroups()

	for _, val := range peerHandler.PeerGroups {
		defer val.Close()
	}

	internal.Log.Info("Starting peer discovery...")

	internal.Log.Infof("Initialization complete!")

	rpc.PeerHandler = &peerHandler
	remote.Handler = &peerHandler
	peerHandler.RunBackground(ctx, h, dht)
}

func main() {
	configFilePath := flag.String("c", "", "path to config file")
	showDoc := flag.Bool("d", false, "show doc")
	testRun := flag.Bool("t", false, "test run")
	verbose := flag.Bool("verbose", false, "log debug info")

	flag.Parse()

	internal.InitLogger()
	internal.SetVerbose(*verbose)

	if *showDoc {
		fmt.Print(GetDocumentation())
		return
	}

	if *configFilePath == "" {
		internal.Log.Warning("No config file given. Use --help to show usage.")
		return
	}

	configFile, err := os.ReadFile(*configFilePath)

	if err != nil {
		internal.Log.Errorf("Error loading config file: %s", err)
		return
	}

	var f []internal.KillSwitchConfig
	err = json.Unmarshal(configFile, &f)

	if err != nil {
		internal.Log.Errorf("Error pasing json file: %s", err)
		return
	}

	ctx := context.Background()

	ctxRemote, _ := context.WithCancel(ctx)
	go runRemoteHandler(ctxRemote)
	time.Sleep(time.Second * 5)

	triggerUpdateChan := make(triggers.TriggerUpdateChan)
	go Observe(triggerUpdateChan)

	for _, cfg := range f {
		trigger, err := triggers.NewTrigger(cfg)
		trigger.TestRun = *testRun

		if err != nil {
			internal.Log.Errorf("%s", err)
			return
		}

		internal.Log.Infof("Registered trigger with name: %s", trigger.Name)
		trigger.Attach(triggerUpdateChan)

		ctxTrigger, cancelTrigger := context.WithCancel(ctx)
		go trigger.Run(ctxTrigger)
		rpc.TriggerList = append(rpc.TriggerList, rpc.TriggerHandlerWithCancel{
			Cancel:  cancelTrigger,
			Trigger: trigger,
		})
	}

	rpc.Serve(*dbPath)
	select {}
}
