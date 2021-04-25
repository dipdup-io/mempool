package models

// MempoolProposal -
type MempoolProposal struct {
	MempoolOperation
	Period    int64  `json:"period"`
	Proposals string `json:"proposal"`
}

// TableName -
func (MempoolProposal) TableName() string {
	return "mempool_proposal"
}
