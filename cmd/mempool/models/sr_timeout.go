package models

// SmartRollupTimeout -
type SmartRollupTimeout struct {
	//nolint
	tableName struct{} `pg:"sr_timeout"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"sr_timeout_source_idx"`
	Rollup       string `json:"rollup,omitempty" index:"sr_timeout_rollup_idx"`
}

// SetMempoolOperation -
func (i *SmartRollupTimeout) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
