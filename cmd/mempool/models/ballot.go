package models

// Ballot -
type Ballot struct {
	//nolint
	tableName struct{} `pg:"ballots"`
	MempoolOperation
	Period int64  `json:"period"`
	Ballot string `json:"ballot"`
}

// SetMempoolOperation -
func (i *Ballot) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
