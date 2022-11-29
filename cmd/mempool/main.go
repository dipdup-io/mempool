package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	libCfg "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/hasura"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
)

type startResult struct {
	cancel  context.CancelFunc
	indexer *Indexer
}

var (
	rootCmd = &cobra.Command{
		Use:   "mempool",
		Short: "DipDup mempool indexer",
	}
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}).Level(zerolog.InfoLevel)

	configPath := rootCmd.PersistentFlags().StringP("config", "c", "dipdup.yml", "path to YAML config file")
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("command line execute")
		return
	}
	if err := rootCmd.MarkFlagRequired("config"); err != nil {
		log.Panic().Err(err).Msg("config command line arg is required")
		return
	}

	var cfg config.Config
	if err := libCfg.Parse(*configPath, &cfg); err != nil {
		log.Err(err).Msg("")
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

			err := startFunc(network, mempool)
			if err == nil {
				return
			}
			log.Err(err).Msg("")

			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					err := startFunc(network, mempool)
					if err == nil {
						return
					}
					log.Err(err).Msg("")
				}
			}
		}(network, mempool)
	}

	<-started

	views, err := createViews(ctx, cfg.Database)
	if err != nil {
		log.Err(err).Msg("")
		cancel()
		return
	}

	if cfg.Hasura != nil {
		t := make([]string, 0)
		for kind := range kinds {
			t = append(t, kind)
		}
		if err := hasura.Create(ctx, hasura.GenerateArgs{
			Config:         cfg.Hasura,
			DatabaseConfig: cfg.Database,
			Views:          views,
			Models:         models.GetModelsBy(t...),
		}); err != nil {
			log.Err(err).Msg("")
			cancel()
			return
		}
	}

	<-signals
	log.Warn().Msg("Trying carefully stopping....")

	for _, indexerCancel := range indexerCancels {
		indexerCancel()
	}

	cancel()

	for _, indexer := range indexers {
		indexer.Close()
	}

	if prometheusService != nil {
		if err := prometheusService.Close(); err != nil {
			log.Err(err).Msg("")
		}
	}

	wg.Wait()

	close(signals)
	close(started)
}

func createViews(ctx context.Context, database libCfg.Database) ([]string, error) {
	files, err := os.ReadDir("views")
	if err != nil {
		return nil, err
	}

	db, err := models.OpenDatabaseConnection(ctx, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	views := make([]string, 0)
	for i := range files {
		if files[i].IsDir() {
			continue
		}

		path := fmt.Sprintf("views/%s", files[i].Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if _, err := db.DB().Exec(string(raw)); err != nil {
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
