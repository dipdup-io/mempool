package models

import (
	"context"
	"time"

	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	pg "github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/rs/zerolog/log"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(ctx context.Context, cfg config.Database, kinds ...string) (db *database.PgGo, err error) {
	db = database.NewPgGo()

	if err := db.Connect(ctx, cfg); err != nil {
		return nil, err
	}

	database.Wait(ctx, db, 5*time.Second)

	data := GetModelsBy(kinds...)
	data = append(data, &State{})

	for i := range data {
		if err := db.DB().WithContext(ctx).Model(data[i]).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			if err := db.Close(); err != nil {
				return nil, err
			}
			return nil, err
		}
	}
	db.DB().AddQueryHook(dbLogger{})

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
	log.Debug().Msgf("%+v", sql)

	return nil
}
