package models

import (
	"context"

	"github.com/uptrace/bun"
)

// Endorsement -
type Endorsement struct {
	bun.BaseModel `bun:"table:endorsements" comment:"endorsement is an operation, which specifies the head of the chain as seen by the endorser of a given slot. The endorser is randomly selected to be included in the block that extends the head of the chain as specified in this operation. A block with more endorsements improves the weight of the chain and increases the likelihood of that chain being the canonical one."`

	MempoolOperation
	Level uint64 `comment:"The height of the block from the genesis block, in which the operation was included." json:"level"`
	Baker string `comment:"Address of the baker who sent the operation."                                         index:"transaction_baker_idx" json:"-"`
}

// EndorsementsWithoutBaker -
func EndorsementsWithoutBaker(ctx context.Context, db bun.IDB, network string, limit, offset int) (endorsements []Endorsement, err error) {
	err = db.NewSelect().Model(&endorsements).
		Where("baker IS NULL").
		Where("network = ?", network).
		Order("level asc").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	return
}
