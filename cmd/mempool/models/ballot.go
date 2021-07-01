package models

// Ballot -
type Ballot struct {
	MempoolOperation
	Period int64  `json:"period"`
	Ballot string `json:"ballot"`
}

// TableName -
func (Ballot) TableName() string {
	return "ballots"
}
