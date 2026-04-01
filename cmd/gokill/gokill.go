package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
	"github.com/k4lipso/gokill/internal/sip"
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
			internal.Log.Infof("Trigger %s initialized", update.Trigger.GetName())
		case triggers.Armed:
			internal.Log.Infof("Trigger %s armed", update.Trigger.GetName())
		case triggers.Firing:
			internal.Log.Infof("Trigger %s firing", update.Trigger.GetName())
		case triggers.Failed:
			internal.Log.Errorf("Trigger %s failed. Reason: %s", update.Trigger.GetName(), update.Error)
		case triggers.Done:
			internal.Log.Infof("Trigger %s done", update.Trigger.GetName())
		}
	}
}

var (
	cfgBaseDir             = flag.String("db", "/etc/gokill", "path to gokill config basedir")
	KeyDirPath             = "/keys"
	AgeKeyPath             = "/age.key"
	Libp2pKeyPath          = "/libp2p.key"
	GokillRemoteConfigPath = "/gokill_remote.json"
)

func runRemoteHandler(ctx context.Context, remoteConfigPath string, ageKeyPath string, libp2pPath string) {
	if ageKeyPath == "" || libp2pPath == "" {
		err := internal.EnsureDirExists(*cfgBaseDir + KeyDirPath)
		if err != nil {
			internal.Log.Error(err.Error())
			return
		}

		if ageKeyPath == "" {
			ageKeyPath = filepath.Join(*cfgBaseDir, KeyDirPath, AgeKeyPath)
		}

		if libp2pPath == "" {
			libp2pPath = filepath.Join(*cfgBaseDir, KeyDirPath, Libp2pKeyPath)
		}
	}

	if remoteConfigPath == "" {
		remoteConfigPath = filepath.Join(*cfgBaseDir, GokillRemoteConfigPath)
	}

	peerHandler, err := remote.CreatePeerHandler(ctx, remoteConfigPath, ageKeyPath, libp2pPath)

	if err != nil {
		internal.Log.Errorf("%s", err)
		return
	}

	remote.Handler = &peerHandler
	peerHandler.RunBackground(ctx)
}

func runSipHandler(ctx context.Context, sipConfigPath string) {
	sipHandler, err := sip.CreateSipHandler(ctx, sipConfigPath)

	if err != nil {
		internal.Log.Errorf("Error during creation of sip handler: %s", err)
		return
	}

	sip.Handler = &sipHandler
	sipHandler.Run(ctx)
}

func main() {
	configFilePath := flag.String("c", "", "path to config file")
	ageKeyPath := flag.String("key-age", "", "optional path to age key")
	libp2pPath := flag.String("key-p2p", "", "optional path to libp2p key")
	remoteConfigPath := flag.String("remote-config", "", "optional path to remote config")
	sipConfigPath := flag.String("sip-config", "", "optional path to sip config")
	showDoc := flag.Bool("d", false, "show doc")
	testRun := flag.Bool("t", false, "test run")
	runRemote := flag.Bool("r", false, "enable remote triggers and actions")
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
	if *runRemote {
		go runRemoteHandler(ctxRemote, *remoteConfigPath, *ageKeyPath, *libp2pPath)
		time.Sleep(time.Second * 5)
	}

	if *sipConfigPath != "" {
		go runSipHandler(ctxRemote, *sipConfigPath)
		time.Sleep(time.Second * 5)
	}

	triggerUpdateChan := make(triggers.TriggerUpdateChan)
	go Observe(triggerUpdateChan)

	for _, cfg := range f {
		trigger, err := triggers.NewTrigger(cfg)

		if err != nil {
			internal.Log.Errorf("%s", err)
			return
		}

		trigger.TestRun = *testRun

		internal.Log.Infof("Registered trigger with name: %s", trigger.Name)
		trigger.Attach(triggerUpdateChan)

		ctxTrigger, cancelTrigger := context.WithCancel(ctx)
		go trigger.Run(ctxTrigger)
		rpc.TriggerList = append(rpc.TriggerList, rpc.TriggerHandlerWithCancel{
			Cancel:  cancelTrigger,
			Trigger: trigger,
		})
	}

	rpc.Serve(*cfgBaseDir)
	select {}
}
