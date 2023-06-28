package models

// SmartRollupExecute -
type SmartRollupExecute struct {
	//nolint
	tableName struct{} `pg:"sr_execute"`

	MempoolOperation
	Fee                int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."` // DISCUSS
	Counter            int64  `pg:",pk" json:"counter,string" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit           int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit       int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Source             string `json:"source,omitempty" index:"sr_execute_source_idx" comment:"Address of the account who has sent the operation."` // DISCUSS
	Rollup             string `json:"rollup" index:"sr_execute_rollup_idx" comment:"Smart rollup to which the operation was sent."`
	CementedCommitment string `json:"cemented_commitment" comment:"Executed commitment."`
	OutputProof        string `json:"output_proof" comment:"Output proof."`
}

// SetMempoolOperation -
func (i *SmartRollupExecute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
