package models

import "github.com/uptrace/bun"

// Reveal -
type Reveal struct {
	bun.BaseModel `bun:"table:reveals" comment:"reveal operation - is used to reveal the public key associated with an account."`
	MempoolOperation
	Source       string `comment:"Address of the account who has sent the operation."                                 index:"reveal_source_idx"                                             json:"source"`
	Fee          int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	PublicKey    string `comment:"Public key of source address."                                                      json:"public_key"`
}
