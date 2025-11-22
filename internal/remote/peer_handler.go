package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mudler/edgevpn/pkg/utils"
	//"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/google/uuid"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"

	multiaddr "github.com/multiformats/go-multiaddr"

	crypto "github.com/libp2p/go-libp2p/core/crypto"
	routed "github.com/libp2p/go-libp2p/p2p/host/routed"

	agelib "filippo.io/age"
	. "github.com/k4lipso/gokill/internal"
)

var (
	Listen  = libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")
	Handler *PeerHandler
)

func SetupLibp2pHost(ctx context.Context, dbPath string) (host host.Host, dht *dht.IpfsDHT, err error) {
	data := dbPath
	keyPath := filepath.Join(data, "key")
	var priv crypto.PrivKey
	_, err = os.Stat(keyPath)
	if os.IsNotExist(err) {
		priv, _, err = crypto.GenerateKeyPair(crypto.Ed25519, 1)
		if err != nil {
			Log.Fatal(err.Error())
		}
		data, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			Log.Fatal(err.Error())
		}
		err = os.WriteFile(keyPath, data, 0400)
		if err != nil {
			Log.Fatal(err.Error())
		}
	} else if err != nil {
		Log.Fatal(err.Error())
	} else {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			Log.Fatal(err.Error())
		}
		priv, err = crypto.UnmarshalPrivateKey(key)
		if err != nil {
			Log.Fatal(err.Error())
		}

	}

	if err != nil {
		Log.Fatal(err.Error())
	}

	host, err = libp2p.New(libp2p.Identity(priv), Listen)

	if err != nil {
		return nil, nil, err
	}

	dht = initDHT(ctx, host)
	host = routed.Wrap(host, dht)

	return host, dht, nil
}

type Peer struct {
	Id               string `json:"Id"`
	Key              string `json:"Key"`
	ConnectionStatus network.Connectedness
}

type PeerGroupConfig struct {
	Name  string `json:"Name"`
	Id    string `json:"Id"`
	Peers []Peer `json:"Peers"`
}

type Config []PeerGroupConfig

type WhitelistConnectionGater struct {
	whitelistedPeers map[peer.ID]struct{}
}

func (wg *WhitelistConnectionGater) InterceptPeerDial(p peer.ID) (allowed bool) {
	//_, allowed = wg.whitelistedPeers[p]
	return true
}

func (wg *WhitelistConnectionGater) InterceptAddrDial(p peer.ID, addr multiaddr.Multiaddr) bool {
	return wg.InterceptPeerDial(p)
}

func (wg *WhitelistConnectionGater) InterceptAccept(conn network.ConnMultiaddrs) bool {
	addr, err := peer.AddrInfoFromP2pAddr(conn.RemoteMultiaddr())

	if err != nil {
		Log.Debugf("Error InterceptAccept: %s\n", err)
		return false
	}

	return wg.InterceptPeerDial(addr.ID)
}

func (wg *WhitelistConnectionGater) InterceptSecured(direction network.Direction, p peer.ID, conn network.ConnMultiaddrs) (allow bool) {
	return wg.InterceptPeerDial(p)
}

func (wg *WhitelistConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return wg.InterceptPeerDial(conn.RemotePeer()), 0
}

func GetTrustedPeers(config []PeerGroupConfig) map[string][]Peer {
	result := make(map[string][]Peer)
	for _, c := range config {
		result[c.Id] = c.Peers
	}

	return result
}

func InitRootNs() {
	//TODO: check if "SharedKeyRegistry" key exists, if not create
}

func PeerFromString(str string) (Peer, error) {
	parts := strings.Split(str, "/")

	if len(parts) != 2 {
		return Peer{}, fmt.Errorf("Invalid Peer String")
	}
	//TODO: validate each part

	return Peer{Id: parts[0], Key: parts[1]}, nil
}

type PeerHandler struct {
	Ctx        context.Context
	Host       host.Host
	PubSub     *pubsub.PubSub
	Key        *agelib.X25519Identity
	Config     []PeerGroupConfig
	PeerGroups map[string]*PeerGroup
	ConfigPath string
}

func (s *PeerHandler) GetSelfPeer() Peer {
	return Peer{
		Id:  s.Host.ID().String(),
		Key: s.Key.Recipient().String(),
	}
}

func (s *PeerHandler) UpdateConfig() {
	Log.Debug("Updating Config...")
	s.recreateConfig()
	s.writeConfig(s.ConfigPath, s.Config)
}

func (s *PeerHandler) recreateConfig() {
	var newCfg []PeerGroupConfig
	for key, val := range s.PeerGroups {
		newCfg = append(newCfg, PeerGroupConfig{
			Name:  key,
			Id:    val.ID,
			Peers: val.TrustedPeers,
		})
	}
	s.Config = newCfg
	//for idx, peerGroupConfig := range s.Config {
	//	s.Config[idx].Peers = s.PeerGroups[peerGroupConfig.Name].TrustedPeers
	//}
}

func (s *PeerHandler) writeConfig(filename string, config []PeerGroupConfig) error {
	jsonData, err := json.Marshal(config)

	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (s *PeerHandler) NewConfig(filename string) ([]PeerGroupConfig, error) {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		err := s.writeConfig(filename, []PeerGroupConfig{
			{
				Name: "root",
				Id:   uuid.New().String(),
				Peers: []Peer{
					{
						Id:  s.Host.ID().String(),
						Key: s.Key.Recipient().String(),
					},
				},
			},
		},
		)

		if err != nil {
			return nil, fmt.Errorf("Could not create config file: %s", err)
		}
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read config file: %s", err)
	}

	var result []PeerGroupConfig
	err = json.Unmarshal(content, &result)

	if err != nil {
		return nil, fmt.Errorf("Could not parse config file: %s", err)
	}

	Log.Infof("Loaded config")
	return result, nil
}

func (s *PeerHandler) GetDefaultPeerGroup(Name string) *PeerGroup {
	return s.PeerGroups["root"]
}

func (s *PeerHandler) InitPeerGroups() {
	peerGroupMap := make(map[string]*PeerGroup)
	Log.Debugf("Init PeerGroups")
	Log.Debugf("Config: %s", s.Config)
	for _, nsCfg := range s.Config {
		ns1, err := CreatePeerGroup(nsCfg.Id, s)

		if err != nil {
			Log.Fatal(err.Error())
		}

		peerGroupMap[nsCfg.Name] = ns1
	}

	s.PeerGroups = peerGroupMap
}

func IsTrustedPeer(ctx context.Context, id peer.ID, peerGroupId string, config []PeerGroupConfig) bool {
	peerMap := GetTrustedPeers(config)

	val, ok := peerMap[peerGroupId]

	if ok {
		for _, v := range val {
			Log.Debugf("Current: %s, Wanted: %s", v.Id, id.String())
			if v.Id == id.String() {
				return true
			}
		}
	}

	return false
}

func (s *PeerHandler) ListPeerGroups() []PeerGroupConfig {
	var result []PeerGroupConfig
	for k, v := range s.PeerGroups {
		result = append(result, PeerGroupConfig{
			Name:  k,
			Id:    v.ID,
			Peers: v.TrustedPeers,
		})
	}

	return result
}

func (s *PeerHandler) DeletePeerGroup(ID string) error {
	ns, ok := s.PeerGroups[ID]

	if !ok {
		Log.Debug("DeletePeerGroup that does not exists")
		return nil
	}

	delete(s.PeerGroups, ID)
	ns.Close()
	s.UpdateConfig()
	return nil
}

func (s *PeerHandler) AddPeerGroup(Name string) (*PeerGroup, error) {
	ns, ok := s.PeerGroups[Name]

	if ok {
		return ns, nil
	}

	result, err := CreatePeerGroup(uuid.New().String(), s)

	if err != nil {
		return nil, err
	}

	result.TrustedPeers = append(result.TrustedPeers, s.GetSelfPeer())
	s.PeerGroups[Name] = result
	s.UpdateConfig()
	return result, nil
}

func (n *PeerHandler) Broadcast(peerGroupName string, msg string) error {
	peerGroup, ok := n.PeerGroups[peerGroupName]

	if !ok {
		return fmt.Errorf("PeerGroup not found.")
	}

	return peerGroup.Broadcast(msg)
}

func (n *PeerHandler) RegisterRemoteTrigger(peerGroupName string, secret string, testSecret string) (chan TriggerMessage, error) {
	peerGroup, ok := n.PeerGroups[peerGroupName]

	if !ok {
		return nil, fmt.Errorf("PeerGroup not found.")
	}

	return peerGroup.RegisterRemoteTrigger(secret, testSecret)
}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}
	//	if err = kademliaDHT.Bootstrap(ctx); err != nil {
	//		panic(err)
	//	}

	return kademliaDHT
}

func (s *PeerHandler) bootstrapPeers(ctx context.Context, h host.Host) {
	Log.Info("Bootstrapping DHT")

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				Log.Debugf("Bootstrap warning: %s", err)
			} else {
				Log.Debugf("Connection established with bootstrap node: %q\n", *peerinfo)
			}
		}()
	}
	wg.Wait()
}

func (s *PeerHandler) RunBackground(ctx context.Context, h host.Host, dht *dht.IpfsDHT) {
	s.bootstrapPeers(ctx, h)
	s.discoverPeers(ctx, h, dht)
	t := utils.NewBackoffTicker(utils.BackoffInitialInterval(2*time.Minute),
		utils.BackoffMaxInterval(6*time.Minute))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			// We announce ourselves to the rendezvous point for all the peers.
			// We have a safeguard of 1 hour to avoid blocking the main loop
			// in case of network issues.
			// The TTL of DHT is by default no longer than 3 hours, so we should
			// be safe by having an entry less than that.
			safeTimeout, cancel := context.WithTimeout(ctx, time.Hour)

			endChan := make(chan struct{})
			go func() {
				s.discoverPeers(safeTimeout, h, dht)
				endChan <- struct{}{}
			}()

			select {
			case <-endChan:
				cancel()
			case <-safeTimeout.Done():
				Log.Error("Timeout while peer discovery")
				cancel()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *PeerHandler) discoverPeers(ctx context.Context, h host.Host, dht *dht.IpfsDHT) error {
	time.Sleep(2 * time.Second)

	for peerGroupName, v := range s.PeerGroups {
		Log.Debugf("Announcing PeerGroup \"%s\" with id: %s", peerGroupName, v.ID)
		routingDiscovery := discovery.NewRoutingDiscovery(dht)
		routingDiscovery.Advertise(ctx, v.ID)

		Log.Debugf("Start peer discovery...")

		timedCtx, cf := context.WithTimeout(ctx, time.Second*120)
		defer cf()

		peerChan, err := routingDiscovery.FindPeers(timedCtx, v.ID)
		if err != nil {
			return err
		}

		for peer := range peerChan {
			if peer.ID == h.ID() || len(peer.Addrs) == 0 {
				continue // No self connection
			}

			if !IsTrustedPeer(timedCtx, peer.ID, v.ID, s.Config) {
				continue // Only conntect to trusted peers
			}

			Log.Debugf("Found peer with id %s", peer.ID.String())
			v.SetPeerConnectionState(peer.ID.String(), h.Network().Connectedness(peer.ID))

			if h.Network().Connectedness(peer.ID) == network.Connected {
				Log.Debugf("Already connected to %s", peer.ID.String())
				continue
			}

			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*120)
			defer cancel()
			err := h.Connect(timeoutCtx, peer)
			if err != nil {
				Log.Debugf("Failed connecting to %s, error: %s\n", peer.ID, err)
			} else {
				Log.Debugf("Connected to: %s", peer.ID)
			}
		}
	}

	Log.Debug("Peer discovery complete")
	return nil
}

func printErr(err error) {
	Log.Errorf("error: %s", err)
}

func ConnectedPeers(h host.Host) []*peer.AddrInfo {
	var pinfos []*peer.AddrInfo
	for _, c := range h.Network().Conns() {
		pinfos = append(pinfos, &peer.AddrInfo{
			ID:    c.RemotePeer(),
			Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
		})
	}
	return pinfos
}
