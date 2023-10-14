package models

import "github.com/uptrace/bun"

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	bun.BaseModel `bun:"register_global_constant" comment:"register_constant operation - is used to register a global constant - Micheline expression that can be reused by multiple smart contracts."`
	MempoolOperation
	Source       string `comment:"Address of the account who has sent the operation."                                 json:"source"`
	Fee          string `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee"`
	Counter      string `comment:"An account nonce which is used to prevent operation replay."                        json:"counter"`
	GasLimit     string `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit"`
	StorageLimit string `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit"`
	Value        JSONB  `bun:"value,type:jsonb"                                                                       comment:"Constant value (Micheline JSON)." json:"value"`
}

// SetMempoolOperation -
func (i *RegisterGlobalConstant) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
