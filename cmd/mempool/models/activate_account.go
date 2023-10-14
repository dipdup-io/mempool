package models

import "github.com/uptrace/bun"

// ActivateAccount -
type ActivateAccount struct {
	bun.BaseModel `bun:"table:activate_account" comment:"activation operation - is used to activate accounts that were recommended allocations of tezos tokens for donations to the Tezos Foundation’s fundraiser."`

	MempoolOperation
	Pkh    string `comment:"Public key hash (Ed25519). Address to activate."                                     json:"pkh"`
	Secret string `comment:"The secret key associated with the key, if available. /^([a-zA-Z0-9][a-zA-Z0-9])*$/" json:"secret"`
}
