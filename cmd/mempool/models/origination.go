package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

// Origination -
type Origination struct {
	bun.BaseModel `bun:"originations" comment:"origination - deployment / contract creation operation."`
	MempoolOperation
	Fee          int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Balance      string `comment:"The contract origination balance (micro tez)."                                      json:"balance"`
	Delegate     string `comment:"Address of the baker (delegate), which was marked as a delegate in the operation."  json:",omitempty"`
	Source       string `comment:"Address of the account who has sent the operation."                                 index:"origination_source_idx"                                        json:"source,omitempty"`
	Script       struct {
		Storage json.RawMessage `json:"storage"`
	} `json:"script" bun:"-"`

	Storage JSONB `bun:",type:jsonb" comment:"Initial contract storage value converted to human-readable JSON." json:"-"`
}

// Fill -
func (mo *Origination) Fill() {
	mo.Storage = JSONB(mo.Script.Storage)
}
