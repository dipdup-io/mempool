package models

// MempoolDelegation -
type MempoolDelegation struct {
	MempoolOperation
	Fee          string `json:"fee"`
	Counter      string `json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	Delegate     string `json:",omitempty"`
}

// TableName -
func (MempoolDelegation) TableName() string {
	return "mempool_delegation"
}
