package models

// SmartRollupExecute -
type SmartRollupExecute struct {
	//nolint
	tableName struct{} `pg:"sr_execute"`

	MempoolOperation
	Fee                int64  `json:"fee,string"`
	Counter            int64  `pg:",pk" json:"counter,string"`
	GasLimit           int64  `json:"gas_limit,string"`
	StorageLimit       int64  `json:"storage_limit,string"`
	Source             string `json:"source,omitempty" index:"sr_execute_source_idx"`
	Rollup             string `json:"rollup" index:"sr_execute_rollup_idx"`
	CementedCommitment string `json:"cemented_commitment"`
	OutputProof        string `json:"output_proof"`
}

// SetMempoolOperation -
func (i *SmartRollupExecute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
