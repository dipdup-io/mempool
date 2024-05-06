package models

import "github.com/uptrace/bun"

// DalPublishCommitment -
type DalPublishCommitment struct {
	bun.BaseModel `bun:"table:dal_publish_commitment"`

	MempoolOperation
	Fee          int64      `comment:"Fee to the baker, produced block, in which the operation was included (micro tez)." json:"fee,string"`
	Counter      int64      `bun:",pk"                                                                                    comment:"An account nonce which is used to prevent operation replay." json:"counter,string"`
	GasLimit     int64      `comment:"A cap on the amount of gas a given operation can consume."                          json:"gas_limit,string"`
	StorageLimit int64      `comment:"A cap on the amount of storage a given operation can consume."                      json:"storage_limit,string"`
	Source       string     `comment:"Address of the account who has sent the operation."                                 index:"sr_refute_source_idx"                                          json:"source,omitempty"`
	SlotHeader   SlotHeader `comment:"Published slot header"                                                              json:"slot_header"`
}

type SlotHeader struct {
	SlotIndex       int    `json:"slot_index"`
	Commitment      string `json:"commitment"`
	CommitmentProof string `json:"commitment_proof"`
}

// SetMempoolOperation -
func (i *DalPublishCommitment) SetMempoolOperation(operaiton MempoolOperation) {
	i.MempoolOperation = operaiton
}
