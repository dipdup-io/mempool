package models

// SmartRollupOriginate -
type SmartRollupOriginate struct {
	//nolint
	tableName struct{} `pg:"sr_originate"`

	MempoolOperation
	Fee              int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter          int64  `pg:",pk" json:"counter,string" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit         int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit     int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Source           string `json:"source,omitempty" index:"sr_originate_source_idx" comment:"Address of the account who has sent the operation."`
	PvmKind          string `json:"pvm_kind" comment:"PVM kind (arith or wasm)."`
	Kernel           string `json:"kernel" comment:"Kernel bytes (hex string)."`
	OriginationProof string `json:"origination_proof" comment:"Origination proof bytes (hex string)."`
	ParametersTy     string `json:"parameters_ty" comment:"Smart rollup parameter type (Micheline JSON)."`
}

// SetMempoolOperation -
func (i *SmartRollupOriginate) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
