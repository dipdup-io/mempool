package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// GasStats -
type GasStats struct {
	bun.BaseModel `bun:"gas_stats"`

	Network        string `bun:",pk"                                                                            comment:"Identifies belonging network." json:"network"`
	Hash           string `bun:",pk"                                                                            comment:"Hash of the operation."        json:"hash"`
	TotalGasUsed   uint64 `bun:"total_gas_used"                                                                 comment:"Total amount of consumed gas." json:"total_gas_used"`
	TotalFee       uint64 `bun:"total_fee"                                                                      comment:"Total amount of fee."          json:"total_fee"`
	UpdatedAt      int64  `comment:"Date of last update in seconds since UNIX epoch."                           json:"updated_at"`
	LevelInMempool uint64 `comment:"Level of the block at which the statistics has been calculated in mempool." json:"level_in_mempool"`
	LevelInChain   uint64 `comment:"Level of the block at which the statistics has been calculated in chain."   json:"level_in_chain"`
}

// BeforeInsert -
func (s *GasStats) BeforeInsert(ctx context.Context) (context.Context, error) {
	s.UpdatedAt = time.Now().Unix()
	return ctx, nil
}

// BeforeUpdate -
func (s *GasStats) BeforeUpdate(ctx context.Context) (context.Context, error) {
	s.UpdatedAt = time.Now().Unix()
	return ctx, nil
}

// Append -
func (s *GasStats) Save(ctx context.Context, db bun.IDB) error {
	if s.TotalGasUsed+s.LevelInChain+s.LevelInMempool == 0 {
		_, err := db.NewInsert().Model(s).Exec(ctx)
		return err
	}

	query := db.NewInsert().Model(s).
		On("CONFLICT (network, hash) DO UPDATE")

	if s.TotalGasUsed > 0 {
		query.Set("total_gas_used = gas_stats.total_gas_used + excluded.total_gas_used")
	}
	if s.TotalFee > 0 {
		query.Set("total_fee = gas_stats.total_fee + excluded.total_fee")
	}
	if s.LevelInChain > 0 {
		query.Set("level_in_chain = excluded.level_in_chain")
	}
	if s.LevelInMempool > 0 {
		query.Set("level_in_mempool = case gas_stats.level_in_mempool when 0 then excluded.level_in_mempool else gas_stats.level_in_mempool end")
	}

	_, err := query.Exec(ctx)
	return err
}

// DeleteOldGasStats -
func DeleteOldGasStats(ctx context.Context, db bun.IDB, timeout uint64) error {
	_, err := db.NewDelete().Model((*GasStats)(nil)).Where("updated_at < ?", time.Now().Unix()-int64(timeout)).Exec(ctx)
	return err
}
