package models

import (
	"encoding/json"
)

// Origination -
type Origination struct {
	//nolint
	tableName struct{} `pg:"originations" comment:"origination - deployment / contract creation operation."`
	MempoolOperation
	Fee          int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter      int64  `json:"counter,string" pg:",pk" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Balance      string `json:"balance" comment:"The contract origination balance (micro tez)."`
	Delegate     string `json:",omitempty" comment:"Address of the baker (delegate), which was marked as a delegate in the operation."`
	Source       string `json:"source,omitempty" index:"origination_source_idx" comment:"Address of the account who has sent the operation."`
	Script       struct {
		Storage json.RawMessage `json:"storage"`
	} `json:"script" pg:"-"`

	Storage JSONB `json:"-" pg:",type:jsonb" comment:"Initial contract storage value converted to human-readable JSON."`
}

// Fill -
func (mo *Origination) Fill() {
	mo.Storage = JSONB(mo.Script.Storage)
}
