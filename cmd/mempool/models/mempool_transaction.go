package models

import "gorm.io/datatypes"

// MempoolTransaction -
type MempoolTransaction struct {
	MempoolOperation
	Source       string         `json:"source"`
	Fee          int64          `json:"fee,string"`
	Counter      int64          `gorm:"primaryKey" json:"counter,string"`
	GasLimit     int64          `json:"gas_limit,string"`
	StorageLimit int64          `json:"storage_limit,string"`
	Amount       string         `json:"amount"`
	Destination  string         `json:"destination"`
	Parameters   datatypes.JSON `json:"parameters,omitempty"`
}

// TableName -
func (MempoolTransaction) TableName() string {
	return "mempool_transaction"
}
