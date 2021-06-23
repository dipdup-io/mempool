package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dipdup-net/go-lib/cmdline"
	libCfg "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/hasura"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	args := cmdline.Parse()
	if args.Help {
		return
	}

	cfg, err := config.Load(args.Config)
	if err != nil {
		log.Error(err)
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	indexers := make(map[string]*Indexer)

	kinds := make(map[string]struct{})

	for network, mempool := range cfg.Mempool.Indexers {
		for _, kind := range mempool.Filters.Kinds {
			kinds[kind] = struct{}{}
		}

		indexer, err := NewIndexer(network, *mempool, cfg.Database, cfg.Mempool.Settings)
		if err != nil {
			log.Error(err)
			return
		}
		indexers[network] = indexer

		if err := indexer.Start(); err != nil {
			log.Error(err)
			return
		}
	}

	views, err := createViews(cfg.Database)
	if err != nil {
		log.Error(err)
		return
	}

	if cfg.Hasura.URL != "" {
		t := make([]string, 0)
		for kind := range kinds {
			t = append(t, kind)
		}
		tables := models.GetModelsBy(t...)
		if err := hasura.Create(cfg.Hasura, cfg.Database, views, tables...); err != nil {
			log.Error(err)
			return
		}
	}

	<-signals
	log.Warn("Trying carefully stopping....")
	for _, indexer := range indexers {
		if err := indexer.Close(); err != nil {
			log.Error(err)
			return
		}
	}

	close(signals)
}

func createViews(database libCfg.Database) ([]string, error) {
	files, err := ioutil.ReadDir("views")
	if err != nil {
		return nil, err
	}

	db, err := models.OpenDatabaseConnection(database)
	if err != nil {
		return nil, err
	}
	sql, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer sql.Close()

	views := make([]string, 0)
	for i := range files {
		if files[i].IsDir() {
			continue
		}

		path := fmt.Sprintf("views/%s", files[i].Name())
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := db.Exec(string(raw)).Error; err != nil {
			return nil, err
		}
		views = append(views, strings.Split(files[i].Name(), ".")[0])
	}

	return views, nil
}
