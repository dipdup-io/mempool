package models

// SmartRollupCement -
type SmartRollupCement struct {
	//nolint
	tableName struct{} `pg:"sr_cement"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"sr_cement_source_idx"`
	Rollup       string `json:"rollup" index:"sr_cement_rollup_idx"`
	Commitment   string `json:"commitment"`
}

// SetMempoolOperation -
func (i *SmartRollupCement) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
