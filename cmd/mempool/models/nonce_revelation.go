package models

// NonceRevelation -
type NonceRevelation struct {
	//nolint
	tableName struct{} `pg:"nonce_revelations"`
	MempoolOperation
	Level int    `json:"level"`
	Nonce string `json:"nonce"`
}

// SetMempoolOperation -
func (i *NonceRevelation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
