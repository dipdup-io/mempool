package models

// MempoolActivateAccount -
type MempoolActivateAccount struct {
	MempoolOperation
	Pkh    string `json:"pkh"`
	Secret string `json:"secret"`
}
