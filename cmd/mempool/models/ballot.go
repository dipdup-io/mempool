package models

// Ballot -
type Ballot struct {
	//nolint
	tableName struct{} `pg:"ballots"`
	MempoolOperation
	Period int64  `json:"period"`
	Ballot string `json:"ballot"`
}
