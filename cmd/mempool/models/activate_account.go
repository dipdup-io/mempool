package models

// ActivateAccount -
type ActivateAccount struct {
	//nolint
	tableName struct{} `pg:"activate_account"`
	MempoolOperation
	Pkh    string `json:"pkh"`
	Secret string `json:"secret"`
}
