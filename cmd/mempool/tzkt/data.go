package tzkt

import (
	"sync"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/api"
	"github.com/dipdup-net/go-lib/tzkt/events"
)

var toNodeKinds = map[string]string{
	api.KindActivation:      node.KindActivation,
	api.KindBallot:          node.KindBallot,
	api.KindDelegation:      node.KindDelegation,
	api.KindDoubleBaking:    node.KindDoubleBaking,
	api.KindDoubleEndorsing: node.KindDoubleEndorsing,
	api.KindEndorsement:     node.KindEndorsement,
	api.KindNonceRevelation: node.KindNonceRevelation,
	api.KindOrigination:     node.KindOrigination,
	api.KindProposal:        node.KindProposal,
	api.KindReveal:          node.KindReveal,
	api.KindTransaction:     node.KindTransaction,
}

// OperationMessage -
type OperationMessage struct {
	Level uint64
	Block string
	Hash  *sync.Map
}

func newOperationMessage() OperationMessage {
	return OperationMessage{
		Hash: new(sync.Map),
	}
}

func (msg *OperationMessage) clear() {
	msg.Hash.Range(func(key, value interface{}) bool {
		msg.Hash.Delete(key)
		return true
	})
	msg.Level = 0
	msg.Block = ""
}

func (msg *OperationMessage) copy() OperationMessage {
	message := newOperationMessage()
	message.Level = msg.Level
	message.Block = msg.Block
	msg.Hash.Range(func(key, value interface{}) bool {
		message.Hash.Store(key, value)
		return true
	})
	return message
}

// BlockMessage -
type BlockMessage struct {
	Hash  string             `json:"hash"`
	Level uint64             `json:"level"`
	Type  events.MessageType `json:"type"`
}
