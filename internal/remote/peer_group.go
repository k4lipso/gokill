package remote

import (
	"context"
	"fmt"
	"time"

	agelib "filippo.io/age"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"

	. "github.com/k4lipso/gokill/internal"
	age "github.com/k4lipso/gokill/internal/age"
)

type PeerGroup struct {
	ID    string
	topic *pubsub.Topic
	//Registry *sharedKeyRegistry
	CancelFunc   context.CancelFunc
	ctx          context.Context
	Key          *agelib.X25519Identity
	TrustedPeers []Peer
}

func (n *PeerGroup) AddPeer(peer Peer) {
	for _, CurrentPeer := range n.TrustedPeers {
		if CurrentPeer.Id == peer.Id && CurrentPeer.Key == peer.Key {
			return
		}
	}

	n.TrustedPeers = append(n.TrustedPeers, peer)
}

func (n *PeerGroup) RemovePeer(peer Peer) {
	var Peers []Peer
	for _, CurrentPeer := range n.TrustedPeers {
		if CurrentPeer.Id == peer.Id && CurrentPeer.Key == peer.Key {
			continue
		}

		Peers = append(Peers, CurrentPeer)
	}

	n.TrustedPeers = Peers
}

func (n *PeerGroup) GetRecipients() []string {
	var result []string

	for _, peer := range n.TrustedPeers {
		result = append(result, peer.Key)
	}

	return result
}

func (n *PeerGroup) Broadcast(msg string) error {
	encryptedMsg, err := age.Encrypt([]byte(msg), n.GetRecipients())
	if err != nil {
		return err
	}

	if err := n.topic.Publish(n.ctx, encryptedMsg); err != nil {
		fmt.Println("### Publish error:", err)
	} else {
		fmt.Println("Sent " + msg)
	}

	return nil
}

func (n *PeerGroup) Close() {
	n.CancelFunc()
}

func streamConsoleTo(ctx context.Context, topic *pubsub.Topic) {
	for {
		s := "Hello World"
		if err := topic.Publish(ctx, []byte(s)); err != nil {
			fmt.Println("### Publish error:", err)
		} else {
			fmt.Println("Sent hello world")
		}

		time.Sleep(10 * time.Second)
	}
}

func printMessagesFrom(ctx context.Context, sub *pubsub.Subscription, peerGroup *PeerGroup) {
	for {
		m, err := sub.Next(ctx)
		if err != nil {
			panic(err)
		}
		msg, err := age.Decrypt(m.Message.Data, peerGroup.Key)

		if err != nil {
			panic(err)
		}

		fmt.Println(m.ReceivedFrom, ": ", string(msg))
	}
}

func CreatePeerGroup(ID string, peerHandler *PeerHandler) (*PeerGroup, error) {
	Log.Infof("Creating PeerGroup %s\n", ID)
	err := peerHandler.PubSub.RegisterTopicValidator(
		ID, //== topicName
		func(ctx context.Context, id peer.ID, msg *pubsub.Message) bool {
			if id == peerHandler.Host.ID() {
				return true
			}

			Log.Debugf("PubSubmsg TOPIC: %s, PEER: %s\n", ID, id)
			trusted := IsTrustedPeer(ctx, id, ID, peerHandler.Config)
			if !trusted {
				Log.Debugf("discarded pubsub message from non trusted source %s\n", id)
			}
			return trusted
		},
	)
	if err != nil {
		Log.Errorf("error registering topic validator: %s", err)
	}

	topic, err := peerHandler.PubSub.Join(ID)
	if err != nil {
		Log.Fatal(err.Error())
		return nil, err
	}

	ctx := context.Background()
	//go streamConsoleTo(ctx, topic)

	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	PeerMap := GetTrustedPeers(peerHandler.Config)
	val, ok := PeerMap[ID]

	if !ok {
		Log.Debug("peerGroup config does not contain any peers")
	}

	//return nil, fmt.Errorf("ERRORRR")
	peerGroup := PeerGroup{ID: ID, topic: topic, ctx: ctx, Key: peerHandler.Key, TrustedPeers: val}
	go printMessagesFrom(ctx, sub, &peerGroup)
	return &peerGroup, nil
}
