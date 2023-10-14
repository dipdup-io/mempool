package models

import "github.com/uptrace/bun"

// SmartRollupExecute -
type SmartRollupExecute struct {
	bun.BaseModel `bun:"sr_execute"`

	MempoolOperation
	Fee                int64  `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter            int64  `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit           int64  `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit       int64  `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source             string `comment:"Address of the account who has sent the operation."                                 index:"sr_execute_source_idx"                                         json:"source,omitempty"`
	Rollup             string `comment:"Smart rollup to which the operation was sent."                                      index:"sr_execute_rollup_idx"                                         json:"rollup"`
	CementedCommitment string `comment:"Executed commitment."                                                               json:"cemented_commitment"`
	OutputProof        string `comment:"Output proof bytes (hex string)."                                                   json:"output_proof"`
}

// SetMempoolOperation -
func (i *SmartRollupExecute) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
