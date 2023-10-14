package models

import "github.com/uptrace/bun"

// NonceRevelation -
type NonceRevelation struct {
	bun.BaseModel `bun:"nonce_revelations" comment:"nonce_revelation operation - are used by the blockchain to create randomness."`
	MempoolOperation
	Level int    `comment:"The height of the block from the genesis block, in which the operation was included." json:"level"`
	Nonce string `comment:"Seed nonce hex."                                                                      json:"nonce"`
}

// SetMempoolOperation -
func (i *NonceRevelation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
