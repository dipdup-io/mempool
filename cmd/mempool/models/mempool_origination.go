package models

import "encoding/json"

// MempoolOrigination -
type MempoolOrigination struct {
	MempoolOperation
	Fee          string `json:"fee"`
	Counter      string `json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	Balance      string `json:"balance"`
	Script       struct {
		Storage json.RawMessage `json:"storage"`
	} `json:"script" gorm:"-"`

	Storage JSON `json:"-"`
}

// TableName -
func (MempoolOrigination) TableName() string {
	return "mempool_origination"
}

// Fill -
func (mo *MempoolOrigination) Fill() {
	mo.Storage = JSON(mo.Script.Storage)
}
