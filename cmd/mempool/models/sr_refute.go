package models

// SmartRollupRefute -
type SmartRollupRefute struct {
	//nolint
	tableName struct{} `pg:"sr_refute"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"sr_refute_source_idx"`
	Opponent     string `json:"opponent,omitempty" index:"sr_refute_opponent_idx"`
	Rollup       string `json:"rollup,omitempty" index:"sr_refute_rollup_idx"`
}

// SetMempoolOperation -
func (i *SmartRollupRefute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
