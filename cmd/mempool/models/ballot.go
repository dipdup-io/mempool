package models

import "github.com/uptrace/bun"

// Ballot -
type Ballot struct {
	bun.BaseModel `bun:"table:ballots" comment:"ballot operation - is used to vote for a proposal in a given voting cycle."`
	MempoolOperation
	Period int64  `comment:"Voting period index, starting from zero, for which the ballot was submitted." json:"period"`
	Ballot string `comment:"Vote, given in the ballot (yay, nay, or pass)."                               json:"ballot"`
}

// SetMempoolOperation -
func (i *Ballot) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
