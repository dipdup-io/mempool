package models

import "github.com/uptrace/bun"

// Transaction -
type Transaction struct {
	bun.BaseModel `bun:"table:transactions"`

	MempoolOperation
	Source       string `comment:"Address of the account who has sent the operation."                                 index:"transaction_source_idx"                                                                   json:"source"`
	Fee          int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay."                            json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Amount       string `comment:"The transaction amount (mutez)."                                                    json:"amount"`
	Destination  string `comment:"Address of the target of the transaction."                                          index:"transaction_destination_idx"                                                              json:"destination"`
	Parameters   JSONB  `bun:",type:jsonb"                                                                            comment:"Transaction parameter, including called entrypoint and value passed to the entrypoint." json:"parameters,omitempty"`
}
