package models

import "github.com/uptrace/bun"

// IncreasePaidStorage -
type IncreasePaidStorage struct {
	bun.BaseModel `bun:"increase_paid_storage"`

	MempoolOperation
}

// SetMempoolOperation -
func (i *IncreasePaidStorage) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
