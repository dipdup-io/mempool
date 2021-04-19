package models

// MempoolEndorsement -
type MempoolEndorsement struct {
	MempoolOperation
	Level uint64 `json:"level"`
}
