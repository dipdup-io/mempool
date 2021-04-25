package models

// MempoolActivateAccount -
type MempoolActivateAccount struct {
	MempoolOperation
	Pkh    string `json:"pkh"`
	Secret string `json:"secret"`
}

// TableName -
func (MempoolActivateAccount) TableName() string {
	return "mempool_activate_account"
}
