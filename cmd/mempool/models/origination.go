package models

import (
	"encoding/json"
)

// Origination -
type Origination struct {
	//nolint
	tableName struct{} `pg:"originations"`
	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `json:"counter,string" pg:",pk" `
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Balance      string `json:"balance"`
	Delegate     string `json:",omitempty"`
	Source       string `json:"source,omitempty" index:"origination_source_idx"`
	Script       struct {
		Storage json.RawMessage `json:"storage"`
	} `json:"script" pg:"-"`

	Storage JSONB `json:"-" pg:"type:jsonb"`
}

// Fill -
func (mo *Origination) Fill() {
	mo.Storage = JSONB(mo.Script.Storage)
}
