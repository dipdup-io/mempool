package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// MempoolOrigination -
type MempoolOrigination struct {
	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `gorm:"primaryKey" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Balance      string `json:"balance"`
	Delegate     string `json:",omitempty"`
	Source       string `json:"source,omitempty" gorm:"origination_source_idx"`
	Script       struct {
		Storage json.RawMessage `json:"storage"`
	} `json:"script" gorm:"-"`

	Storage datatypes.JSON `json:"-"`
}

// TableName -
func (MempoolOrigination) TableName() string {
	return "mempool_origination"
}

// Fill -
func (mo *MempoolOrigination) Fill() {
	mo.Storage = datatypes.JSON(mo.Script.Storage)
}
