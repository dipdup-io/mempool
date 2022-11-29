package models

// UpdateConsensusKey -
type UpdateConsensusKey struct {
	//nolint
	tableName struct{} `pg:"update_consensus_key"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source" index:"update_consensus_key_source_idx"`
	Pk           string `json:"pk"`
}

// SetMempoolOperation -
func (i *UpdateConsensusKey) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
