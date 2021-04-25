package models

// MempoolReveal -
type MempoolReveal struct {
	MempoolOperation
	Source       string `json:"source"`
	Fee          string `json:"fee"`
	Counter      string `json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	PublicKey    string `json:"public_key"`
}

// TableName -
func (MempoolReveal) TableName() string {
	return "mempool_reveal"
}
