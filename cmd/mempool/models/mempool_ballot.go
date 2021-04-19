package models

// MempoolBallot -
type MempoolBallot struct {
	MempoolOperation
	Period int64  `json:"period"`
	Ballot string `json:"ballot"`
}
