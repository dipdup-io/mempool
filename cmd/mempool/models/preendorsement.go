package models

// Preendorsement -
type Preendorsement struct {
	//nolint
	tableName struct{} `pg:"preendorsements"`
	MempoolOperation
}
