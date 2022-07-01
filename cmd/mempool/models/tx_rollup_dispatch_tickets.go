package models

// TxRollupDispatchTickets -
type TxRollupDispatchTickets struct {
	//nolint
	tableName struct{} `pg:"tx_rollup_dispatch_tickets"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"tx_rollup_dispatch_tickets_source_idx"`
	TxRollup     string `json:"tx_rollup,omitempty" index:"tx_rollup_dispatch_tickets_rollup_idx"`
}
