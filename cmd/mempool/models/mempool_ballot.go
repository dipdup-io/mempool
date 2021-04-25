package models

// MempoolBallot -
type MempoolBallot struct {
	MempoolOperation
	Period int64  `json:"period"`
	Ballot string `json:"ballot"`
}

// TableName -
func (MempoolBallot) TableName() string {
	return "mempool_ballot"
}
