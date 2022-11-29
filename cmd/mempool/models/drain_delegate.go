package models

// DelegateDrain -
type DelegateDrain struct {
	//nolint
	tableName struct{} `pg:"drain_delegate"`

	MempoolOperation
	ConsensusKey int64  `json:"consensus_key"`
	Delegate     string `json:"delegate"`
	Destination  string `json:"destination"`
}

// SetMempoolOperation -
func (i *DelegateDrain) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
