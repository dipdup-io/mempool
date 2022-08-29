package models

// Delegation -
type Delegation struct {
	//nolint
	tableName struct{} `pg:"delegations"`
	MempoolOperation
	Fee          int64  `json:"fee,string"`
	Counter      int64  `json:"counter,string" pg:",pk"`
	GasLimit     int64  `json:"gas_limit,string"`
	StorageLimit int64  `json:"storage_limit,string"`
	Delegate     string `json:",omitempty"`
	Source       string `json:"source,omitempty" index:"delegation_source_idx"`
}

// SetMempoolOperation -
func (i *Delegation) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
