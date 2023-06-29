package models

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	//nolint
	tableName struct{} `pg:"register_global_constant" comment:"register_constant operation - is used to register a global constant - Micheline expression that can be reused by multiple smart contracts."`
	MempoolOperation
	Source       string `json:"source" comment:"Address of the account who has sent the operation."`
	Fee          string `json:"fee" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter      string `json:"counter" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     string `json:"gas_limit" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit string `json:"storage_limit" comment:"A cap on the amount of storage a given operation can consume."`
	Value        JSONB  `json:"value"  pg:"value,type:jsonb" comment:"Constant value (Micheline JSON)."`
}

// SetMempoolOperation -
func (i *RegisterGlobalConstant) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
