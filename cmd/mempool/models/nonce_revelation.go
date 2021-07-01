package models

// NonceRevelation -
type NonceRevelation struct {
	MempoolOperation
	Level int    `json:"level"`
	Nonce string `json:"nonce"`
}

// TableName -
func (NonceRevelation) TableName() string {
	return "nonce_revelations"
}
