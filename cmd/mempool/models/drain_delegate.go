package models

import "github.com/uptrace/bun"

// DelegateDrain -
type DelegateDrain struct {
	bun.BaseModel `bun:"table:drain_delegate"`

	MempoolOperation
	ConsensusKey int64  `comment:"Consensus key that was used to sign Drain." json:"consensus_key"`
	Delegate     string `comment:"Address of the drained delegate."           json:"delegate"`
	Destination  string `comment:"Address of the recipient account."          json:"destination"`
}

// SetMempoolOperation -
func (i *DelegateDrain) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
