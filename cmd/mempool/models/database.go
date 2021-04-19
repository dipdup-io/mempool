package models

import (
	"github.com/dipdup-net/mempool/internal/config"
	"github.com/dipdup-net/mempool/internal/state"
	"gorm.io/gorm"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(cfg config.Database, kinds ...string) (*gorm.DB, error) {
	db, err := state.OpenConnection(cfg.Kind, cfg.Path)
	if err != nil {
		return nil, err
	}

	data := GetModelsBy(kinds...)
	data = append(data, &state.State{})

	if err := db.AutoMigrate(data...); err != nil {
		sql, err := db.DB()
		if err == nil {
			if err := sql.Close(); err != nil {
				return nil, err
			}
		}
		return nil, err
	}
	return db, nil
}
