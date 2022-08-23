package models

// TxRollupCommit -
type TxRollupCommit struct {
	//nolint
	tableName struct{} `pg:"tx_rollup_commit"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"tx_rollup_commit_source_idx"`
	Rollup       string `json:"rollup,omitempty" index:"tx_rollup_commit_rollup_idx"`
}

// SetMempoolOperation -
func (i *TxRollupCommit) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
