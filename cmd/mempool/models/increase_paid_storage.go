package models

// IncreasePaidStorage -
type IncreasePaidStorage struct {
	//nolint
	tableName struct{} `pg:"increase_paid_storage"`

	MempoolOperation
}

// SetMempoolOperation -
func (i *IncreasePaidStorage) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
