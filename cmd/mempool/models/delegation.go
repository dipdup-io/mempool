package models

// Delegation -
type Delegation struct {
	//nolint
	tableName struct{} `pg:"delegations" comment:"Type of the operation, delegation - is used to delegate funds to a delegate (an implicit account registered as a baker)."`
	MempoolOperation
	Fee          int64  `json:"fee,string" comment:"Fee to a baker, produced block, in which the operation was included."`
	Counter      int64  `json:"counter,string" pg:",pk" comment:"An account nonce which is used to prevent operation replay."`
	GasLimit     int64  `json:"gas_limit,string" comment:"A cap on the amount of gas a given operation can consume."`
	StorageLimit int64  `json:"storage_limit,string" comment:"A cap on the amount of storage a given operation can consume."`
	Delegate     string `json:",omitempty" comment:"Address of the delegate to which the operation was sent. null if there is no new delegate (an un-delegation operation)."`
	Source       string `json:"source,omitempty" index:"delegation_source_idx" comment:"Address of the delegated account."`
}

// SetMempoolOperation -
func (i *Delegation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
