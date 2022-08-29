package tzkt

import (
	"sync"
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/data"
	"github.com/dipdup-net/go-lib/tzkt/events"
)

var toNodeKinds = map[string]string{
	data.KindActivation:               node.KindActivation,
	data.KindBallot:                   node.KindBallot,
	data.KindDelegation:               node.KindDelegation,
	data.KindDoubleBaking:             node.KindDoubleBaking,
	data.KindDoubleEndorsing:          node.KindDoubleEndorsing,
	data.KindEndorsement:              node.KindEndorsement,
	data.KindNonceRevelation:          node.KindNonceRevelation,
	data.KindOrigination:              node.KindOrigination,
	data.KindProposal:                 node.KindProposal,
	data.KindReveal:                   node.KindReveal,
	data.KindTransaction:              node.KindTransaction,
	data.KindRegisterGlobalConstant:   node.KindRegisterGlobalConstant,
	data.KindRollupDispatchTickets:    node.KindTxRollupDispatchTickets,
	data.KindRollupFinalizeCommitment: node.KindTxRollupFinalizeCommitment,
	data.KindRollupReturnBond:         node.KindTxRollupReturnBond,
	data.KindRollupSubmitBatch:        node.KindTxRollupSubmitBatch,
	data.KindTransferTicket:           node.KindTransferTicket,
	data.KindTxRollupCommit:           node.KindTxRollupCommit,
	data.KindTxRollupOrigination:      node.KindTxRollupOrigination,
	data.KindTxRollupRejection:        node.KindTxRollupRejection,
	data.KindTxRollupRemoveCommitment: node.KindTxRollupRemoveCommitment,
	data.KindSetDepositsLimit:         node.KindSetDepositsLimit,
	data.KindIncreasePaidStorage:      node.KindIncreasePaidStorage,
	data.KindVdfRevelation:            node.KindVdfRevelation,
}

var toTzKTKinds = map[string]string{
	node.KindActivation:                 data.KindActivation,
	node.KindBallot:                     data.KindBallot,
	node.KindDelegation:                 data.KindDelegation,
	node.KindDoubleBaking:               data.KindDoubleBaking,
	node.KindDoubleEndorsing:            data.KindDoubleEndorsing,
	node.KindEndorsement:                data.KindEndorsement,
	node.KindNonceRevelation:            data.KindNonceRevelation,
	node.KindOrigination:                data.KindOrigination,
	node.KindProposal:                   data.KindProposal,
	node.KindReveal:                     data.KindReveal,
	node.KindTransaction:                data.KindTransaction,
	node.KindRegisterGlobalConstant:     data.KindRegisterGlobalConstant,
	node.KindTxRollupDispatchTickets:    data.KindRollupDispatchTickets,
	node.KindTxRollupFinalizeCommitment: data.KindRollupFinalizeCommitment,
	node.KindTxRollupReturnBond:         data.KindRollupReturnBond,
	node.KindTxRollupSubmitBatch:        data.KindRollupSubmitBatch,
	node.KindTransferTicket:             data.KindTransferTicket,
	node.KindTxRollupCommit:             data.KindTxRollupCommit,
	node.KindTxRollupOrigination:        data.KindTxRollupOrigination,
	node.KindTxRollupRejection:          data.KindTxRollupRejection,
	node.KindTxRollupRemoveCommitment:   data.KindTxRollupRemoveCommitment,
	node.KindSetDepositsLimit:           data.KindSetDepositsLimit,
	node.KindIncreasePaidStorage:        data.KindIncreasePaidStorage,
	node.KindVdfRevelation:              data.KindVdfRevelation,
}

// OperationMessage -
type OperationMessage struct {
	Level     uint64
	Block     string
	Timestamp time.Time
	Hash      *sync.Map
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
	msg.Timestamp = time.Now().UTC()
}

func (msg *OperationMessage) copy() OperationMessage {
	message := newOperationMessage()
	message.Level = msg.Level
	message.Block = msg.Block
	message.Timestamp = msg.Timestamp
	msg.Hash.Range(func(key, value interface{}) bool {
		message.Hash.Store(key, value)
		return true
	})
	return message
}

// BlockMessage -
type BlockMessage struct {
	Hash      string             `json:"hash"`
	Level     uint64             `json:"level"`
	Type      events.MessageType `json:"type"`
	Timestamp time.Time
}
