package models

// Proposal -
type Proposal struct {
	//nolint
	tableName struct{} `pg:"proposals"`
	MempoolOperation
	Period    int64  `json:"period"`
	Proposals string `json:"proposal"`
}
