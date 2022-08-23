package models

// TransferTicket -
type TransferTicket struct {
	//nolint
	tableName struct{} `pg:"transfer_ticket"`

	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `pg:",pk" json:"counter,string"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Source       string `json:"source,omitempty" index:"transfer_ticket_source_idx"`
}

// SetMempoolOperation -
func (i *TransferTicket) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
