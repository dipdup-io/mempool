package models

import "github.com/go-pg/pg/v10"

// Endorsement -
type Endorsement struct {
	//nolint
	tableName struct{} `pg:"endorsements" comment:"endorsement is an operation, which specifies the head of the chain as seen by the endorser of a given slot. The endorser is randomly selected to be included in the block that extends the head of the chain as specified in this operation. A block with more endorsements improves the weight of the chain and increases the likelihood of that chain being the canonical one."`
	MempoolOperation
	Level uint64 `json:"level" comment:"The height of the block from the genesis block, in which the operation was included."`
	Baker string `json:"-" index:"transaction_baker_idx" comment:"Address of the baker who sent the operation."` // DISCUSS
}

// EndorsementsWithoutBaker -
func EndorsementsWithoutBaker(db pg.DBI, network string, limit, offset int) (endorsements []Endorsement, err error) {
	err = db.Model(&endorsements).
		Where("baker IS NULL").
		Where("network = ?", network).
		Order("level asc").
		Limit(limit).
		Offset(offset).
		Select()
	return
}
