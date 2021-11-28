package models

import (
	"context"

	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/state"
	"gorm.io/gorm"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(ctx context.Context, cfg config.Database, kinds ...string) (*gorm.DB, error) {
	db, err := state.OpenConnection(ctx, cfg)
	if err != nil {
		return nil, err
	}

	sql, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.Kind == config.DBKindSqlite {
		sql.SetMaxOpenConns(1)
	}

	data := GetModelsBy(kinds...)
	data = append(data, &state.State{})

	if err := db.AutoMigrate(data...); err != nil {
		if err := sql.Close(); err != nil {
			return nil, err
		}
		return nil, err
	}
	return db, nil
}
