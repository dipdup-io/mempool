package models

// Transaction -
type Transaction struct {
	//nolint
	tableName struct{} `pg:"transactions"`

	MempoolOperation
	Source       string `json:"source" index:"transaction_source_idx" comment:"Address of the account who has sent the operation."`      // DISCUSS
	Fee          int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."` // DISCUSS
	Counter      int64  `json:"counter,string" pg:",pk" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Amount       string `json:"amount" comment:"The transaction amount (mutez)."`
	Destination  string `json:"destination" index:"transaction_destination_idx" comment:"Address of the target of the transaction."`
	Parameters   JSONB  `json:"parameters,omitempty" pg:",type:jsonb" comment:"Transaction parameter, including called entrypoint and value passed to the entrypoint."`
}
