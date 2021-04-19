package models

// MempoolNonceRevelation -
type MempoolNonceRevelation struct {
	MempoolOperation
	Level int    `json:"level"`
	Nonce string `json:"nonce"`
}
