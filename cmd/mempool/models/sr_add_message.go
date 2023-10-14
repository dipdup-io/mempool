package models

import "github.com/uptrace/bun"

// SetDepositsLimit -
type SmartRollupAddMessage struct {
	bun.BaseModel `bun:"table:sr_add_messages"`

	MempoolOperation
	Fee          int64    `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64    `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64    `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64    `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source       string   `comment:"Address of the account who has sent the operation."                                 index:"set_deposits_limit_source_idx"                                 json:"source,omitempty"`
	Message      []string `comment:"Messages added to the smart rollup inbox (Array of hex strings)."                   json:"message"`
}

// SetMempoolOperation -
func (i *SmartRollupAddMessage) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
