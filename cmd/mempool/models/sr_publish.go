package models

import "github.com/uptrace/bun"

// SmartRollupPublish -
type SmartRollupPublish struct {
	bun.BaseModel `bun:"sr_publish"`

	MempoolOperation
	Fee          int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source       string `comment:"Address of the account who has sent the operation."                                 index:"sr_publish_source_idx"                                         json:"source,omitempty"`
	Rollup       string `comment:"Smart rollup to which the operation was sent."                                      index:"sr_publish_bond_rollup_idx"                                    json:"rollup,omitempty"`
}

// SetMempoolOperation -
func (i *SmartRollupPublish) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
