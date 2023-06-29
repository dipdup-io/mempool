package models

// TxRollupDispatchTickets -
type TxRollupDispatchTickets struct {
	//nolint
	tableName struct{} `pg:"tx_rollup_dispatch_tickets"`

	MempoolOperation
	Fee          int64  `json:"fee,string" comment:"Fee to the baker, produced block, in which the operation was included (micro tez)."`
	Counter      int64  `pg:",pk" json:"counter,string" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Source       string `json:"source,omitempty" index:"tx_rollup_dispatch_tickets_source_idx" comment:"Address of the account who has sent the operation."`
	TxRollup     string `json:"tx_rollup,omitempty" index:"tx_rollup_dispatch_tickets_rollup_idx" comment:"Address of the rollup to which the operation was sent."`
}

// SetMempoolOperation -
func (i *TxRollupDispatchTickets) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
