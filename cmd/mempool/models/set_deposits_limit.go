package models

// SetDepositsLimit -
type SetDepositsLimit struct {
	//nolint
	tableName struct{} `pg:"set_deposits_limit"`

	MempoolOperation
	Fee          int64   `json:"fee,string"`
	Counter      int64   `pg:",pk" json:"counter,string"`
	GasLimit     int64   `json:"gas_limit,string"`
	StorageLimit int64   `json:"storage_limit,string"`
	Source       string  `json:"source,omitempty" index:"set_deposits_limit_source_idx"`
	Limit        *string `json:"limit,omitempty"`
}

// SetMempoolOperation -
func (i *SetDepositsLimit) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
