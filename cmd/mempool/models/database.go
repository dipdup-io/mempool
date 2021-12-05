package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/dipdup-net/go-lib/config"
	pg "github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	log "github.com/sirupsen/logrus"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(ctx context.Context, cfg config.Database, kinds ...string) (db *pg.DB, err error) {
	if cfg.Kind != config.DBKindPostgres {
		return nil, errors.New("unsupported database type")
	}
	if cfg.Path != "" {
		opt, err := pg.ParseURL(cfg.Path)
		if err != nil {
			return nil, err
		}
		db = pg.Connect(opt)
	} else {
		db = pg.Connect(&pg.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			User:     cfg.User,
			Password: cfg.Password,
			Database: cfg.Database,
		})
	}

	if err = db.Ping(ctx); err != nil {
		return nil, err
	}

	data := GetModelsBy(kinds...)
	data = append(data, &State{})

	for i := range data {
		if err := db.WithContext(ctx).Model(data[i]).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			if err := db.Close(); err != nil {
				return nil, err
			}
			return nil, err
		}
	}
	db.AddQueryHook(dbLogger{})

	return db, nil
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	raw, err := q.FormattedQuery()
	if err != nil {
		return err
	}
	sql := string(raw)
	log.Debugf("%+v", sql)

	return nil
}
