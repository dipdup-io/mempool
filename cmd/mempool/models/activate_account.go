package models

// ActivateAccount -
type ActivateAccount struct {
	//nolint
	tableName struct{} `pg:"activate_account" comment:"Type of the operation, activation - is used to activate accounts that were recommended allocations of tezos tokens for donations to the Tezos Foundation’s fundraiser."`
	MempoolOperation
	Pkh    string `json:"pkh" comment:"Public key hash (Ed25519). Address to activate."`
	Secret string `json:"secret" comment:"The secret key associated with the key, if available."`
}
