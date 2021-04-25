package models

// MempoolNonceRevelation -
type MempoolNonceRevelation struct {
	MempoolOperation
	Level int    `json:"level"`
	Nonce string `json:"nonce"`
}

// TableName -
func (MempoolNonceRevelation) TableName() string {
	return "mempool_nonce_revelation"
}
