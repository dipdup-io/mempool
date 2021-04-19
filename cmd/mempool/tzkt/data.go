package tzkt

import (
	"sync"
	"time"

	"github.com/dipdup-net/go-lib/node"
	events "github.com/dipdup-net/tzktevents"
)

const (
	KindTransaction     = "transaction"
	KindOrigination     = "origination"
	KindEndorsement     = "endorsement"
	KindBallot          = "ballot"
	KindProposal        = "proposal"
	KindDoubleBaking    = "double_baking"
	KindDoubleEndorsing = "double_endorsing"
	KindActivation      = "activation"
	KindNonceRevelation = "nonce_revelation"
	KindDelegation      = "delegation"
	KindReveal          = "reveal"
)

var toNodeKinds = map[string]string{
	KindActivation:      node.KindActivation,
	KindBallot:          node.KindBallot,
	KindDelegation:      node.KindDelegation,
	KindDoubleBaking:    node.KindDoubleBaking,
	KindDoubleEndorsing: node.KindDoubleEndorsing,
	KindEndorsement:     node.KindEndorsement,
	KindNonceRevelation: node.KindNonceRevelation,
	KindOrigination:     node.KindOrigination,
	KindProposal:        node.KindProposal,
	KindReveal:          node.KindReveal,
	KindTransaction:     node.KindTransaction,
}

// Operation -
type Operation struct {
	ID    uint64 `json:"id"`
	Level uint64 `json:"level"`
	Hash  string `json:"hash"`
	Kind  string `json:"type"`
	Block string `json:"block"`
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

// Block -
type Block struct {
	Level         uint64    `json:"level"`
	Hash          string    `json:"hash"`
	Timestamp     time.Time `json:"timestamp"`
	Proto         int64     `json:"proto"`
	Priority      int64     `json:"priority"`
	Validations   int64     `json:"validations"`
	Deposit       int64     `json:"deposit"`
	Reward        int64     `json:"reward"`
	Fees          int64     `json:"fees"`
	NonceRevealed bool      `json:"nonceRevealed"`
	Baker         struct {
		Alias   string `json:"alias"`
		Address string `json:"address"`
	} `json:"baker"`
}

// Head -
type Head struct {
	Level        uint64    `json:"level"`
	Hash         string    `json:"hash"`
	Protocol     string    `json:"protocol"`
	Timestamp    time.Time `json:"timestamp"`
	VotingEpoch  int64     `json:"votingEpoch"`
	VotingPeriod int64     `json:"votingPeriod"`
	KnownLevel   int64     `json:"knownLevel"`
	LastSync     time.Time `json:"lastSync"`
	Synced       bool      `json:"synced"`
	QuoteLevel   int64     `json:"quoteLevel"`
	QuoteBtc     float64   `json:"quoteBtc"`
	QuoteEur     float64   `json:"quoteEur"`
	QuoteUsd     float64   `json:"quoteUsd"`
}
