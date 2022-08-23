package models

// TxRollupReturnBond -
type TxRollupReturnBond struct {
	//nolint
	tableName struct{} `pg:"tx_rollup_return_bond"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"ttx_rollup_return_bond_source_idx"`
}

// SetMempoolOperation -
func (i *TxRollupReturnBond) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
