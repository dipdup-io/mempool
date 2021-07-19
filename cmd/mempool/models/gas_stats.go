package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GasStats -
type GasStats struct {
	Network        string `gorm:"primaryKey" json:"network"`
	Hash           string `gorm:"primaryKey" json:"hash"`
	TotalGasUsed   uint64 `gorm:"column:total_gas_used" json:"total_gas_used"`
	TotalFee       uint64 `gorm:"column:total_fee" json:"total_fee"`
	UpdatedAt      int    `gorm:"autoUpdateTime" json:"updated_at"`
	LevelInMempool uint64 `json:"level_in_mempool"`
	LevelInChain   uint64 `json:"level_in_chain"`
}

// TableName -
func (GasStats) TableName() string {
	return "gas_stats"
}

// Append -
func (s *GasStats) Save(tx *gorm.DB) error {
	assignments := make(map[string]interface{})

	if s.TotalGasUsed > 0 {
		assignments["total_gas_used"] = gorm.Expr("gas_stats.total_gas_used + ?", s.TotalGasUsed)
	}
	if s.TotalFee > 0 {
		assignments["total_fee"] = gorm.Expr("gas_stats.total_fee + ?", s.TotalFee)
	}
	if s.LevelInChain > 0 {
		assignments["level_in_chain"] = s.LevelInChain
	}
	if s.LevelInMempool > 0 {
		assignments["level_in_mempool"] = gorm.Expr("case gas_stats.level_in_mempool when 0 then ? else gas_stats.level_in_mempool end", s.LevelInMempool)
	}

	if s.TotalGasUsed == 0 && s.LevelInChain == 0 && s.LevelInMempool == 0 {
		return tx.Save(s).Error
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network"},
			{Name: "hash"},
		},
		DoUpdates: clause.Assignments(assignments),
	}).Create(s).Error
}

// DeleteOldGasStats -
func DeleteOldGasStats(db *gorm.DB, timeout uint64) error {
	return db.Where("updated_at < ?", time.Now().Unix()-int64(timeout)).Delete(&GasStats{}).Error
}
