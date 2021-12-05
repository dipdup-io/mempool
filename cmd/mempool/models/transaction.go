package models

// Transaction -
type Transaction struct {
	//nolint
	tableName struct{} `pg:"transactions"`
	MempoolOperation
	Source       string `json:"source" index:"transaction_source_idx"`
	Fee          int64  `json:"fee,string"`
	Counter      int64  `json:"counter,string" pg:",pk"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Amount       string `json:"amount"`
	Destination  string `json:"destination" index:"transaction_destination_idx"`
	Parameters   JSONB  `json:"parameters,omitempty" pg:",type:jsonb"`
}
