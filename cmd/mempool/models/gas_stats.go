package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

// GasStats -
type GasStats struct {
	//nolint
	tableName      struct{} `pg:"gas_stats"`
	Network        string   `pg:",pk" json:"network"`
	Hash           string   `pg:",pk" json:"hash"`
	TotalGasUsed   uint64   `pg:"total_gas_used" json:"total_gas_used"`
	TotalFee       uint64   `pg:"total_fee" json:"total_fee"`
	UpdatedAt      int64    `json:"updated_at"`
	LevelInMempool uint64   `json:"level_in_mempool"`
	LevelInChain   uint64   `json:"level_in_chain"`
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
func (s *GasStats) Save(db pg.DBI) error {
	if s.TotalGasUsed == 0 && s.LevelInChain == 0 && s.LevelInMempool == 0 {
		_, err := db.Model(&s).Insert()
		return err
	}

	query := db.Model(s).OnConflict("(network, hash) DO UPDATE")

	var set strings.Builder
	if s.TotalGasUsed > 0 {
		if _, err := set.WriteString(fmt.Sprintf("total_gas_used = gas_stats.total_gas_used + %d", s.TotalGasUsed)); err != nil {
			return err
		}
	}
	if s.TotalFee > 0 {
		if set.Len() > 0 {
			if err := set.WriteByte(','); err != nil {
				return err
			}
		}
		if _, err := set.WriteString(fmt.Sprintf("total_fee = gas_stats.total_fee + %d", s.TotalFee)); err != nil {
			return err
		}
	}
	if s.LevelInChain > 0 {
		if set.Len() > 0 {
			if err := set.WriteByte(','); err != nil {
				return err
			}
		}
		if _, err := set.WriteString(fmt.Sprintf("level_in_chain = %d", s.LevelInChain)); err != nil {
			return err
		}
	}
	if s.LevelInMempool > 0 {
		if set.Len() > 0 {
			if err := set.WriteByte(','); err != nil {
				return err
			}
		}
		if _, err := set.WriteString(fmt.Sprintf("level_in_mempool = case gas_stats.level_in_mempool when 0 then %d else gas_stats.level_in_mempool end", s.LevelInMempool)); err != nil {
			return err
		}
	}

	if set.Len() > 0 {
		query.Set(set.String())
	}

	_, err := query.Insert()
	return err
}

// DeleteOldGasStats -
func DeleteOldGasStats(db pg.DBI, timeout uint64) error {
	_, err := db.Model((*GasStats)(nil)).Where("updated_at < ?", time.Now().Unix()-int64(timeout)).Delete()
	return err
}
