package models

import "github.com/uptrace/bun"

// Proposal -
type Proposal struct {
	bun.BaseModel `bun:"table:proposals" comment:"proposal operation - is used by bakers (delegates) to submit and/or upvote proposals to amend the protocol."`
	MempoolOperation
	Period    int64  `comment:"Voting period index (starting from zero) for which the proposal was submitted (upvoted)." json:"period"`
	Proposals string `bun:"proposals"                                                                                    comment:"Information about the submitted (upvoted) proposal." json:"proposal"`
}
