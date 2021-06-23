package models

// MempoolDelegation -
type MempoolDelegation struct {
	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `gorm:"primaryKey" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Delegate     string `json:",omitempty"`
	Source       string `json:"source,omitempty"`
}

// TableName -
func (MempoolDelegation) TableName() string {
	return "mempool_delegation"
}
