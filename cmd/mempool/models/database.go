package models

import (
	"log"
	"os"
	"time"

	"github.com/dipdup-net/mempool/internal/config"
	"github.com/dipdup-net/mempool/internal/state"
	"github.com/pkg/errors"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(cfg config.Database, kinds ...string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Kind {
	case DBKindSqlite:
		dialector = sqlite.Open(cfg.Path)
	case DBKindPostgres:
		dialector = postgres.Open(cfg.Path)
	case DBKindMysql:
		dialector = mysql.Open(cfg.Path)
	case DBKindClickHouse:
		dialector = clickhouse.Open(cfg.Path)
	case DBKindSqlServer:
		dialector = sqlserver.Open(cfg.Path)
	default:
		return nil, errors.Errorf("Unsupported database %s", cfg.Kind)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	data := GetModelsBy(kinds...)
	data = append(data, &state.State{})

	return db, db.AutoMigrate(data...)
}
