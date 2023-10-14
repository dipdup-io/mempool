package models

import "github.com/uptrace/bun"

// Preendorsement -
type Preendorsement struct {
	bun.BaseModel `bun:"table:preendorsements"`
	MempoolOperation
}

// SetMempoolOperation -
func (i *Preendorsement) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
