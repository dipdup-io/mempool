package models

// SmartRollupRefute -
type SmartRollupRefute struct {
	//nolint
	tableName struct{} `pg:"sr_refute"`

	MempoolOperation
	Fee          int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter      int64  `pg:",pk" json:"counter,string" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Source       string `json:"source,omitempty" index:"sr_refute_source_idx" comment:"Address of the account who has sent the operation."`
	Opponent     string `json:"opponent,omitempty" index:"sr_refute_opponent_idx" comment:"Address of the opponent, who was accused in publishing a wrong commitment."`
	Rollup       string `json:"rollup,omitempty" index:"sr_refute_rollup_idx" comment:"Smart rollup to which the operation was sent."`
}

// SetMempoolOperation -
func (i *SmartRollupRefute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
