package models

// Preendorsement -
type Preendorsement struct {
	//nolint
	tableName struct{} `pg:"preendorsements"`
	MempoolOperation
}

// SetMempoolOperation -
func (i *Preendorsement) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
