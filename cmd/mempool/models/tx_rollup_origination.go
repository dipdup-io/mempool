package models

// TxRollupOrigination -
type TxRollupOrigination struct {
	//nolint
	tableName struct{} `pg:"tx_rollup_origination"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"ttx_rollup_origination_source_idx"`
}

// SetMempoolOperation -
func (i *TxRollupOrigination) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
