package models

// VdfRevelation -
type VdfRevelation struct {
	//nolint
	tableName struct{} `pg:"vdf_revelation"`

	MempoolOperation
}

// SetMempoolOperation -
func (i *VdfRevelation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
