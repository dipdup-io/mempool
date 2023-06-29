package models

// Reveal -
type Reveal struct {
	//nolint
	tableName struct{} `pg:"reveals" comment:"reveal operation - is used to reveal the public key associated with an account."`
	MempoolOperation
	Source       string `json:"source" index:"reveal_source_idx" comment:"Address of the account who has sent the operation."`
	Fee          int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter      int64  `json:"counter,string" pg:",pk" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	PublicKey    string `json:"public_key" comment:"Public key of source address."`
}
