package models

// MempoolReveal -
type MempoolReveal struct {
	MempoolOperation
	Source       string `json:"source" gorm:"reveal_source_idx"`
	Fee          int64  `json:"fee,string"`
	Counter      int64  `gorm:"primaryKey" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	PublicKey    string `json:"public_key"`
}

// TableName -
func (MempoolReveal) TableName() string {
	return "mempool_reveal"
}
