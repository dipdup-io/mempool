package models

// DelegateDrain -
type DelegateDrain struct {
	//nolint
	tableName struct{} `pg:"drain_delegate"`

	MempoolOperation
	ConsensusKey int64  `json:"consensus_key" comment:"Consensus key that was used to sign Drain."`
	Delegate     string `json:"delegate" comment:"Address of the drained delegate."`
	Destination  string `json:"destination" comment:"Address of the recipient account."`
}

// SetMempoolOperation -
func (i *DelegateDrain) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
