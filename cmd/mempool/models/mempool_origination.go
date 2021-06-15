package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// MempoolOrigination -
type MempoolOrigination struct {
	MempoolOperation
	Fee          string `json:"fee"`
	Counter      string `gorm:"primaryKey" json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	Balance      string `json:"balance"`
	Delegate     string `json:",omitempty"`
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
