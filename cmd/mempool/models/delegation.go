package models

import "github.com/uptrace/bun"

// Delegation -
type Delegation struct {
	bun.BaseModel `bun:"delegations" comment:"delegation operation - is used to delegate funds to a delegate (an implicit account registered as a baker)."`
	MempoolOperation
	Fee          int64  `comment:"Fee to a baker, produced block, in which the operation was included."                                                    json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                                                         comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                                                               json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                                                           json:"storage_limit,string"`
	Delegate     string `comment:"Address of the delegate to which the operation was sent. null if there is no new delegate (an un-delegation operation)." json:",omitempty"`
	Source       string `comment:"Address of the delegated account."                                                                                       index:"delegation_source_idx"                                         json:"source,omitempty"`
}

// SetMempoolOperation -
func (i *Delegation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
