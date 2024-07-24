package rpc

import (
	"time"
	"net"
	"net/rpc"
	"net/http"

  "os"
  "os/signal"
  "syscall"

	"github.com/google/uuid"

	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/triggers"
)

var TriggerList []*triggers.TriggerHandler
var DisabledTriggers []*triggers.TriggerHandler

type TriggerInfo struct {
	Config internal.KillSwitchConfig
	TimeStarted time.Time
	TimeFired time.Time
	Id uuid.UUID
}

type Query int

func (t *Query) DisableTrigger(id *string, success *bool) error {
	var result bool
	result = false

	//delete trigger from triggerlist, create new one in disabledtriggers
	for i := len(TriggerList) - 1; i >= 0; i-- {
		if TriggerList[i].Id.String() == *id {
			internal.Log.Debugf("Disabling Trigger with id: %s", *id)
			newTrig := TriggerList[i]
			newTrig.WrappedTrigger.Enable(false)
			DisabledTriggers = append(DisabledTriggers, newTrig)
			TriggerList = append(TriggerList[:i], TriggerList[i+1:]...)
			result = true
		}
	}

	*success = result
	return nil
}

func (t *Query) ActiveTriggers(_ *int, reply *[]TriggerInfo) error {
	internal.Log.Debug("RPC Request: Query::ActiveTriggers")

	var result []TriggerInfo
	for _, trigger := range TriggerList {
		triggerInfo := TriggerInfo{
			Config: trigger.Config,
			TimeStarted: trigger.TimeStarted,
			TimeFired: trigger.TimeFired,
			Id: trigger.Id,
		}

		result = append(result, triggerInfo)
	}

	*reply = result
	return nil
}

func Serve() {
	query := new(Query)
	rpc.Register(query)
	rpc.HandleHTTP()
	l, err := net.Listen("unix", "/tmp/rpc_test.socket")

	if err != nil {
		internal.Log.Errorf("Error while listening on unix socket: %s", err)
	}


	go http.Serve(l, nil)

	sigc := make(chan os.Signal, 1)
  signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
  func(ln net.Listener, c chan os.Signal) {
    sig := <-c
    internal.Log.Infof("Caught signal %s: shutting down.", sig)
    ln.Close()
    os.Exit(0)
  }(l, sigc)
}

func Receive() (*rpc.Client, error) {
	client, err := rpc.DialHTTP("unix", "/tmp/rpc_test.socket")

	if err != nil {
		internal.Log.Errorf("Cant connect to RPC server: %s", err)
	}

	return client, err
}
