package models

import "github.com/uptrace/bun"

// SmartRollupOriginate -
type SmartRollupOriginate struct {
	bun.BaseModel `bun:"sr_originate"`

	MempoolOperation
	Fee              int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter          int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit         int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit     int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source           string `comment:"Address of the account who has sent the operation."                                 index:"sr_originate_source_idx"                                       json:"source,omitempty"`
	PvmKind          string `comment:"PVM kind (arith or wasm)."                                                          json:"pvm_kind"`
	Kernel           string `comment:"Kernel bytes (hex string)."                                                         json:"kernel"`
	OriginationProof string `comment:"Origination proof bytes (hex string)."                                              json:"origination_proof"`
	ParametersTy     string `comment:"Smart rollup parameter type (Micheline JSON)."                                      json:"parameters_ty"`
}

// SetMempoolOperation -
func (i *SmartRollupOriginate) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
