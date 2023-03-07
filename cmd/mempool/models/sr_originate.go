package models

// SmartRollupOriginate -
type SmartRollupOriginate struct {
	//nolint
	tableName struct{} `pg:"sr_originate"`

	MempoolOperation
	Fee              int64  `json:"fee,string"`
	Counter          int64  `pg:",pk" json:"counter,string"`
	GasLimit         int64  `json:"gas_limit,string"`
	StorageLimit     int64  `json:"storage_limit,string"`
	Source           string `json:"source,omitempty" index:"sr_originate_source_idx"`
	PvmKind          string `json:"pvm_kind"`
	Kernel           string `json:"kernel"`
	OriginationProof string `json:"origination_proof"`
	ParametersTy     string `json:"parameters_ty"`
}

// SetMempoolOperation -
func (i *SmartRollupOriginate) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
