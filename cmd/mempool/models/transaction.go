package models

import "gorm.io/datatypes"

// Transaction -
type Transaction struct {
	MempoolOperation
	Source       string         `json:"source" gorm:"transaction_source_idx"`
	Fee          int64          `json:"fee,string"`
	Counter      int64          `gorm:"primaryKey" json:"counter,string"`
	GasLimit     int64          `json:"gas_limit,string"`
	StorageLimit int64          `json:"storage_limit,string"`
	Amount       string         `json:"amount"`
	Destination  string         `json:"destination" gorm:"transaction_destination_idx"`
	Parameters   datatypes.JSON `json:"parameters,omitempty"`
}

// TableName -
func (Transaction) TableName() string {
	return "transactions"
}