package models

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	//nolint
	tableName struct{} `pg:"register_global_constant" comment:"register_constant operation - is used to register a global constant - Micheline expression that can be reused by multiple smart contracts."`
	MempoolOperation
	Source       string `json:"source"` // DISCUSS: This is sender address or constant address?
	Fee          string `json:"fee"`    // DISCUSS
	Counter      string `json:"counter" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     string `json:"gas_limit" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit string `json:"storage_limit" comment:"A cap on the amount of storage a given operation can consume."`
	Value        JSONB  `json:"value"  pg:"value,type:jsonb" comment:"Constant value."`
}

// SetMempoolOperation -
func (i *RegisterGlobalConstant) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
