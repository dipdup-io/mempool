package models

import pg "github.com/go-pg/pg/v10"

// Endorsement -
type Endorsement struct {
	//nolint
	tableName struct{} `pg:"endorsements"`
	MempoolOperation
	Level uint64 `json:"level"`
	Baker string `json:"-" index:"transaction_baker_idx"`
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
