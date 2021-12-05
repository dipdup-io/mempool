package models

// Reveal -
type Reveal struct {
	//nolint
	tableName struct{} `pg:"reveals"`
	MempoolOperation
	Source       string `json:"source" index:"reveal_source_idx"`
	Fee          int64  `json:"fee,string"`
	Counter      int64  `json:"counter,string" pg:",pk" `
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	PublicKey    string `json:"public_key"`
}
