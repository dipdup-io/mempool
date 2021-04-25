package models

// MempoolTransaction -
type MempoolTransaction struct {
	MempoolOperation
	Source       string `json:"source"`
	Fee          string `json:"fee"`
	Counter      string `json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	Amount       string `json:"amount"`
	Destination  string `json:"destination"`
	Parameters   JSON   `json:"parameters,omitempty"`
}

// TableName -
func (MempoolTransaction) TableName() string {
	return "mempool_transaction"
}
