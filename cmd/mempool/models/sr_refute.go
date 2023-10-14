package models

import "github.com/uptrace/bun"

// SmartRollupRefute -
type SmartRollupRefute struct {
	bun.BaseModel `bun:"table:sr_refute"`

	MempoolOperation
	Fee          int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source       string `comment:"Address of the account who has sent the operation."                                 index:"sr_refute_source_idx"                                          json:"source,omitempty"`
	Opponent     string `comment:"Address of the opponent, who was accused in publishing a wrong commitment."         index:"sr_refute_opponent_idx"                                        json:"opponent,omitempty"`
	Rollup       string `comment:"Smart rollup to which the operation was sent."                                      index:"sr_refute_rollup_idx"                                          json:"rollup,omitempty"`
}

// SetMempoolOperation -
func (i *SmartRollupRefute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
