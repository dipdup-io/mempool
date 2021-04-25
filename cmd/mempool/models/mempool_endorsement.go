package models

// MempoolEndorsement -
type MempoolEndorsement struct {
	MempoolOperation
	Level uint64 `json:"level"`
}

// TableName -
func (MempoolEndorsement) TableName() string {
	return "mempool_endorsement"
}
