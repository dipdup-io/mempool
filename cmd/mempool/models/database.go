package models

import (
	"context"
	"time"

	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// OpenDatabaseConnection -
func OpenDatabaseConnection(ctx context.Context, cfg config.Database, kinds ...string) (db *database.Bun, err error) {
	db = database.NewBun()

	if err := db.Connect(ctx, cfg); err != nil {
		return nil, err
	}

	database.Wait(ctx, db, 5*time.Second)

	db.DB().AddQueryHook(new(logQueryHook))

	data := GetModelsBy(kinds...)
	data = append(data, &database.State{})

	for i := range data {
		if _, err := db.DB().NewCreateTable().Model(data[i]).IfNotExists().Exec(ctx); err != nil {
			if err := db.Close(); err != nil {
				return nil, err
			}
			return nil, err
		}
	}

	if err := database.MakeComments(ctx, db, data...); err != nil {
		return nil, err
	}

	return db, nil
}

type logQueryHook struct{}

// BeforeQuery -
func (h *logQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	event.StartTime = time.Now()
	return ctx
}

// AfterQuery -
func (h *logQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if event.Err != nil {
		log.Trace().Msgf("[%d mcs] %s : %s", time.Since(event.StartTime).Microseconds(), event.Err.Error(), event.Query)
	} else {
		log.Trace().Msgf("[%d mcs] %s", time.Since(event.StartTime).Microseconds(), event.Query)
	}
}
