package models

// Proposal -
type Proposal struct {
	//nolint
	tableName struct{} `pg:"proposals" comment:"proposal operation - is used by bakers (delegates) to submit and/or upvote proposals to amend the protocol."`
	MempoolOperation
	Period    int64  `json:"period" comment:"Voting period index (starting from zero) for which the proposal was submitted (upvoted)."`
	Proposals string `json:"proposal" pg:"proposals" comment:"Information about the submitted (upvoted) proposal."` // DISCUSS
}
