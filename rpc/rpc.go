package rpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"github.com/k4lipso/gokill/actions"
	"github.com/k4lipso/gokill/internal"
	"github.com/k4lipso/gokill/internal/remote"
	"github.com/k4lipso/gokill/triggers"
)

var TriggerList []*triggers.TriggerHandler
var DisabledTriggers []*triggers.TriggerHandler

type TriggerInfo struct {
	Config      internal.KillSwitchConfig
	TimeStarted time.Time
	TimeFired   time.Time
	Id          uuid.UUID
	Active      bool
	State       triggers.TriggerState
}

func (t TriggerInfo) Title() string       { return t.Config.Name }
func (t TriggerInfo) Description() string { return t.Config.Type }
func (t TriggerInfo) FilterValue() string { return t.Config.Name }

type Query int

func (t *Query) EnableTrigger(id *string, success *bool) error {
	var result bool
	result = false

	//delete trigger from triggerlist, create new one in disabledtriggers
	for i := len(TriggerList) - 1; i >= 0; i-- {
		if TriggerList[i].Id.String() == *id {
			internal.Log.Debugf("Enabling Trigger with id: %s", *id)
			TriggerList[i].WrappedTrigger.Enable(true)

			if TriggerList[i].Running == false {
				go TriggerList[i].Listen()
			}

			result = true
		}
	}

	*success = result
	return nil
}

func (t *Query) TestAction(conf internal.ActionConfig, result *error) error {
	internal.Log.Infof("Action Test requested. Type: %s", conf.Type)

	actionChan := make(actions.ActionResultChan)
	action, err := actions.NewSingleAction(conf, actionChan)

	if err != nil {
		internal.Log.Errorf("Error during action test: %s", err)
		*result = err
		return err
	}

	go action.DryExecute()
	err = <-actionChan

	if err != nil {
		internal.Log.Errorf("Error during action test: %s", err)
	}

	return err
}

func (t *Query) DisableTrigger(id *string, success *bool) error {
	var result bool
	result = false

	for i := len(TriggerList) - 1; i >= 0; i-- {
		if TriggerList[i].Id.String() == *id {
			internal.Log.Debugf("Disabling Trigger with id: %s", *id)
			TriggerList[i].WrappedTrigger.Enable(false)
			result = true
		}
	}

	*success = result
	return nil
}

func (t *Query) LoadedTriggers(_ *int, reply *[]TriggerInfo) error {
	internal.Log.Debug("RPC Request: Query::LoadedTriggers")

	var result []TriggerInfo
	for _, trigger := range TriggerList {
		triggerInfo := TriggerInfo{
			Config:      trigger.Config,
			TimeStarted: trigger.TimeStarted,
			TimeFired:   trigger.TimeFired,
			Id:          trigger.Id,
			Active:      trigger.WrappedTrigger.IsEnabled(),
			State:       trigger.State,
		}

		result = append(result, triggerInfo)
	}

	*reply = result
	return nil
}

var PeerHandler *remote.PeerHandler

type PeerGroupService struct {
	PeerGroup string
	Service   string
}

type PeerGroupPeer struct {
	PeerGroup string
	Peer      string
}

func (t *Query) Broadcast(peerGroup *PeerGroupService, _ *string) error {
	val, ok := PeerHandler.PeerGroups[peerGroup.PeerGroup]

	if !ok {
		return fmt.Errorf("PeerGroup does not exist")
	}

	return val.Broadcast(peerGroup.Service)
}

func (t *Query) GetOwnPeerId(_ *int, result *string) error {
	*result = PeerHandler.Host.ID().String()
	return nil
}

func (t *Query) GetPeerString(_ *int, result *string) error {
	*result = PeerHandler.Host.ID().String() + "/" + PeerHandler.Key.Recipient().String()
	return nil
}

func (t *Query) AddPeer(np *PeerGroupPeer, success *bool) error {
	peerGroup := np.PeerGroup
	val, ok := PeerHandler.PeerGroups[peerGroup]

	if !ok {
		return fmt.Errorf("PeerGroup does not exist")
	}

	peer, err := remote.PeerFromString(np.Peer)

	if err != nil {
		internal.Log.Infof("Error parsing peer string: %s\n", err)
		*success = false
		return err
	}

	val.AddPeer(peer)
	*success = true
	PeerHandler.UpdateConfig()
	return nil
}

func (t *Query) DeletePeer(np *PeerGroupPeer, success *bool) error {
	peerGroup := np.PeerGroup
	val, ok := PeerHandler.PeerGroups[peerGroup]

	if !ok {
		return fmt.Errorf("PeerGroup does not exist")
	}

	peer, err := remote.PeerFromString(np.Peer)

	if err != nil {
		internal.Log.Infof("Error parsing peer string: %s\n", err)
		*success = false
		return err
	}

	val.RemovePeer(peer)
	*success = true
	PeerHandler.UpdateConfig()
	return nil
}

func (t *Query) AddPeerGroup(peerGroup *string, _ *int) error {
	_, err := PeerHandler.AddPeerGroup(*peerGroup)
	return err
}

func (t *Query) DeletePeerGroup(peerGroup *string, _ *int) error {
	err := PeerHandler.DeletePeerGroup(*peerGroup)
	return err
}

func (t *Query) ListPeerGroups(_ *int, reply *[]remote.PeerGroupConfig) error {
	*reply = PeerHandler.ListPeerGroups()
	return nil
}

func Serve(path string) {
	query := new(Query)
	rpc.Register(query)
	rpc.HandleHTTP()
	l, err := net.Listen("unix", path+"/rpc_test.socket")

	if err != nil {
		internal.Log.Errorf("Error while listening on unix socket: %s\n", err)
	}

	go http.Serve(l, nil)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		internal.Log.Infof("Caught signal %s: shutting down.\n", sig)
		ln.Close()
		os.Exit(0)
	}(l, sigc)
}

func Receive(path string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("unix", path+"/rpc_test.socket")

	if err != nil {
		internal.Log.Errorf("Cant connect to RPC server: %s\n", err)
	}

	return client, err
}
