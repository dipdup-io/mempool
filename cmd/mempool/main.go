package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dipdup-io/workerpool"
	"github.com/grafana/pyroscope-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	libCfg "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/go-lib/hasura"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/profiler"
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
		log.Err(err).Msg("parse config")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	notifyCtx, notifyCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCancel()

	var prscp *pyroscope.Profiler
	if cfg.Profiler != nil && cfg.Profiler.Server != "" {
		p, err := profiler.New(cfg.Profiler, "indexer")
		if err != nil {
			log.Err(err).Msg("create profiler")
			return
		}
		prscp = p
	}

	var prometheusService *prometheus.Service
	if cfg.Prometheus != nil {
		prometheusService = prometheus.NewService(cfg.Prometheus)
		registerPrometheusMetrics(prometheusService)
		prometheusService.Start()
	}

	kinds := make(map[string]struct{})
	filters := make([]string, 0)
	for _, mempool := range cfg.Mempool.Indexers {
		for _, kind := range mempool.Filters.Kinds {
			if _, ok := kinds[kind]; !ok {
				kinds[kind] = struct{}{}
				filters = append(filters, kind)
			}
		}
	}

	db, err := models.OpenDatabaseConnection(ctx, cfg.Database, filters...)
	if err != nil {
		log.Err(err).Msg("open database connection")
		return
	}

	g := workerpool.NewGroup()
	indexerCancels := make(map[string]context.CancelFunc)
	indexers := make(map[string]*Indexer)

	startFunc := func(ctx context.Context, network string, mempool *config.Indexer) error {
		result, err := startIndexer(ctx, network, cfg, mempool, db, prometheusService)
		if err != nil {
			return err
		}

		indexers[network] = result.indexer
		indexerCancels[network] = result.cancel
		log.Info().Str("network", network).Msg("indexer started")
		return nil
	}

	runFunc := func(ctx context.Context, network string, mempool *config.Indexer) {
		err := startFunc(ctx, network, mempool)
		if err == nil {
			return
		}
		log.Err(err).Msg("start indexer")

		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := startFunc(ctx, network, mempool)
				if err == nil {
					return
				}
				log.Err(err).Msg("start indexer")
			}
		}
	}

	for network, mempool := range cfg.Mempool.Indexers {
		log.Info().Str("network", network).Msg("running indexer...")
		g.GoCtx(ctx, func(ctx context.Context) {
			runFunc(ctx, network, mempool)
		})
		time.Sleep(time.Millisecond * 100)
	}

	g.Wait()

	views, err := createViews(ctx, cfg.Database)
	if err != nil {
		log.Err(err).Msg("creating views")
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
			log.Err(err).Msg("hasura.Create")
			cancel()
			return
		}
	}

	<-notifyCtx.Done()
	log.Info().Msg("Trying carefully stopping....")

	for _, indexerCancel := range indexerCancels {
		indexerCancel()
	}

	cancel()

	for _, indexer := range indexers {
		indexer.Close()
	}

	if prometheusService != nil {
		if err := prometheusService.Close(); err != nil {
			log.Err(err).Msg("stopping prometheus")
		}
	}

	if prscp != nil {
		if err := prscp.Stop(); err != nil {
			log.Panic().Err(err).Msg("stopping pyroscope")
		}
	}
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

func startIndexer(ctx context.Context, network string, cfg config.Config, mempool *config.Indexer, db *database.Bun, prometheusService *prometheus.Service) (startResult, error) {
	var result startResult

	indexerCtx, cancel := context.WithCancel(ctx)
	indexer, err := NewIndexer(indexerCtx, network, *mempool, db, cfg.Mempool.Settings, prometheusService)
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
