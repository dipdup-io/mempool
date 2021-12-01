package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
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

type startResult struct {
	cancel  context.CancelFunc
	indexer *Indexer
}

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

	var prometheusService *prometheus.Service
	if cfg.Prometheus != nil {
		prometheusService = prometheus.NewService(cfg.Prometheus)
		registerPrometheusMetrics(prometheusService)
		prometheusService.Start()
	}

	var wg sync.WaitGroup
	started := make(chan struct{}, len(cfg.Mempool.Indexers))
	indexerCancels := make(map[string]context.CancelFunc)
	kinds := make(map[string]struct{})
	indexers := make(map[string]*Indexer)

	for network, mempool := range cfg.Mempool.Indexers {
		for _, kind := range mempool.Filters.Kinds {
			kinds[kind] = struct{}{}
		}

		startFunc := func(network string, mempool *config.Indexer) error {
			result, err := startIndexer(ctx, network, cfg, mempool, prometheusService)
			if err != nil {
				return err
			}

			indexers[network] = result.indexer
			indexerCancels[network] = result.cancel
			started <- struct{}{}
			return nil
		}

		wg.Add(1)
		go func(network string, mempool *config.Indexer) {
			defer wg.Done()

			if err := startFunc(network, mempool); err == nil {
				return
			}
			log.Error(err)

			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := startFunc(network, mempool); err == nil {
						return
					}
					log.Error(err)
				}
			}
		}(network, mempool)
	}

	<-started

	views, err := createViews(ctx, cfg.Database)
	if err != nil {
		log.Error(err)
		cancel()
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
			cancel()
			return
		}
	}

	<-signals
	log.Warn("Trying carefully stopping....")

	for _, indexerCancel := range indexerCancels {
		indexerCancel()
	}

	cancel()

	for _, indexer := range indexers {
		indexer.Close()
	}

	if prometheusService != nil {
		if err := prometheusService.Close(); err != nil {
			log.Error(err)
		}
	}

	wg.Wait()

	close(signals)
	close(started)
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

func startIndexer(ctx context.Context, network string, cfg config.Config, mempool *config.Indexer, prometheusService *prometheus.Service) (startResult, error) {
	var result startResult

	indexerCtx, cancel := context.WithCancel(ctx)
	indexer, err := NewIndexer(indexerCtx, network, *mempool, cfg.Database, cfg.Mempool.Settings, prometheusService)
	if err != nil {
		cancel()
		return result, err
	}
	result.indexer = indexer

	if err := indexer.Start(indexerCtx); err != nil {
		cancel()
		return result, err
	}
	result.cancel = cancel
	return result, nil
}
