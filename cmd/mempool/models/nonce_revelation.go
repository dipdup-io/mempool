package models

// NonceRevelation -
type NonceRevelation struct {
	//nolint
	tableName struct{} `pg:"nonce_revelations" comment:"nonce_revelation operation - are used by the blockchain to create randomness."`
	MempoolOperation
	Level int    `json:"level" comment:"The height of the block from the genesis block, in which the operation was included."`
	Nonce string `json:"nonce" comment:"Seed nonce hex."`
}

// SetMempoolOperation -
func (i *NonceRevelation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
