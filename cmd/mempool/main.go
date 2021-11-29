package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dipdup-net/go-lib/cmdline"
	libCfg "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/hasura"
	"github.com/dipdup-net/go-lib/prometheus"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	indexers := make(map[string]*Indexer)

	kinds := make(map[string]struct{})

	var prometheusService *prometheus.Service
	if cfg.Prometheus != nil {
		prometheusService = prometheus.NewService(cfg.Prometheus)
		registerPrometheusMetrics(prometheusService)
		prometheusService.Start()
	}

	indexerCancels := make(map[string]context.CancelFunc)
	for network, mempool := range cfg.Mempool.Indexers {
		for _, kind := range mempool.Filters.Kinds {
			kinds[kind] = struct{}{}
		}

		go func(network string, mempool *config.Indexer) {
			indexerCancel, err := startIndexer(ctx, network, cfg, mempool, prometheusService)
			if err != nil {
				log.Error(err)
			} else {
				indexerCancels[network] = indexerCancel
				return
			}

			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cancelFunc, err := startIndexer(ctx, network, cfg, mempool, prometheusService)
					if err != nil {
						log.Error(err)
					} else {
						indexerCancels[network] = cancelFunc
						return
					}
				}
			}
		}(network, mempool)
	}

	views, err := createViews(ctx, cfg.Database)
	if err != nil {
		log.Error(err)
		return
	}

	if cfg.Hasura != nil {
		t := make([]string, 0)
		for kind := range kinds {
			t = append(t, kind)
		}
		tables := models.GetModelsBy(t...)
		if err := hasura.Create(ctx, cfg.Hasura, cfg.Database, views, tables...); err != nil {
			log.Error(err)
			return
		}
	}

	<-signals
	log.Warn("Trying carefully stopping....")

	for network, indexerCancel := range indexerCancels {
		log.Infof("stopping %s indexer...", network)
		indexerCancel()
	}

	for _, indexer := range indexers {
		if err := indexer.Close(); err != nil {
			log.Error(err)
			return
		}
	}

	if prometheusService != nil {
		if err := prometheusService.Close(); err != nil {
			log.Error(err)
		}
	}

	close(signals)
}

func createViews(ctx context.Context, database libCfg.Database) ([]string, error) {
	files, err := ioutil.ReadDir("views")
	if err != nil {
		return nil, err
	}

	db, err := models.OpenDatabaseConnection(ctx, database)
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

func startIndexer(ctx context.Context, network string, cfg config.Config, mempool *config.Indexer, prometheusService *prometheus.Service) (context.CancelFunc, error) {
	indexerCtx, cancel := context.WithCancel(ctx)
	indexer, err := NewIndexer(indexerCtx, network, *mempool, cfg.Database, cfg.Mempool.Settings, prometheusService)
	if err != nil {
		cancel()
		return nil, err
	}

	if err := indexer.Start(indexerCtx); err != nil {
		cancel()
		return nil, err
	}
	return cancel, nil
}
