package models

import "github.com/uptrace/bun"

// VdfRevelation -
type VdfRevelation struct {
	bun.BaseModel `bun:"vdf_revelation"`

	MempoolOperation
}

// SetMempoolOperation -
func (i *VdfRevelation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
