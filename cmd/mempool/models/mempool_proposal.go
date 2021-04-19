package models

// MempoolProposal -
type MempoolProposal struct {
	MempoolOperation
	Period    int64  `json:"period"`
	Proposals string `json:"proposal"`
}
