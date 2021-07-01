package models

// Proposal -
type Proposal struct {
	MempoolOperation
	Period    int64  `json:"period"`
	Proposals string `json:"proposal"`
}

// TableName -
func (Proposal) TableName() string {
	return "proposals"
}
