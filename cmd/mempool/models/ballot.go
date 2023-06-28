package models

// Ballot -
type Ballot struct {
	//nolint
	tableName struct{} `pg:"ballots" comment:"ballot operation - is used to vote for a proposal in a given voting cycle."`
	MempoolOperation
	Period int64  `json:"period" comment:"Voting period index, starting from zero, for which the ballot was submitted."` // DISCUSS
	Ballot string `json:"ballot" comment:"Vote, given in the ballot (yay, nay, or pass)."`
}

// SetMempoolOperation -
func (i *Ballot) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
