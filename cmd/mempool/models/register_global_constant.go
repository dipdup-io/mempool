package models

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	//nolint
	tableName struct{} `pg:"register_global_constant"`
	MempoolOperation
	Source       string `json:"source"`
	Fee          string `json:"fee"`
	Counter      string `json:"counter"`
	GasLimit     string `json:"gas_limit"`
	StorageLimit string `json:"storage_limit"`
	Value        JSONB  `json:"value"  pg:"value,type:jsonb"`
}

// SetMempoolOperation -
func (i *RegisterGlobalConstant) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
