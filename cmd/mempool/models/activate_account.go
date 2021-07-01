package models

// ActivateAccount -
type ActivateAccount struct {
	MempoolOperation
	Pkh    string `json:"pkh"`
	Secret string `json:"secret"`
}

// TableName -
func (ActivateAccount) TableName() string {
	return "activate_account"
}
