package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dipdup-io/workerpool"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"

	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/receiver"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
)

// Indexer -
type Indexer struct {
	db               *database.Bun
	tzkt             *tzkt.TzKT
	mempool          *receiver.Receiver
	prom             *prometheus.Service
	branches         *BlockQueue
	cache            *Cache
	rights           *ccache.Cache
	delegates        *CachedDelegates
	state            *database.State
	logger           zerolog.Logger
	filters          config.Filters
	endorsements     chan *models.Endorsement
	network          string
	indexName        string
	chainID          string
	keepInChain      uint64
	keepOperations   uint64
	gasStatsLifetime uint64
	hasManager       bool

	g workerpool.Group
}

// NewIndexer -
func NewIndexer(ctx context.Context, network string, indexerCfg config.Indexer, db *database.Bun, settings config.Settings, prom *prometheus.Service) (*Indexer, error) {
	rpc := node.NewMainRPC(indexerCfg.DataSource.RPC.Struct().URL)
	constants, err := rpc.Constants(ctx, "head")
	if err != nil {
		return nil, err
	}
	delay := constants.MinimalBlockDelay
	if delay == 0 {
		if len(constants.TimeBetweenBlocks) == 0 {
			return nil, errors.Errorf("Empty time_between_blocks in node response: %s", network)
		}
		delay = constants.TimeBetweenBlocks[0]
	}

	head, err := rpc.Header(ctx, "head")
	if err != nil {
		return nil, err
	}

	memInd, err := receiver.New(indexerCfg.DataSource.URL(), network, db,
		receiver.WithPrometheus(prom),
		receiver.WithBlockTime(delay),
	)
	if err != nil {
		return nil, err
	}

	expiredAfter := settings.ExpiredAfter
	if expiredAfter == 0 {
		metadata, err := rpc.Metadata(ctx, "head")
		if err != nil {
			return nil, err
		}
		expiredAfter = metadata.MaxOperationsTTL
	}

	gasStatsLifetime := settings.GasStatsLifetime
	if gasStatsLifetime == 0 {
		gasStatsLifetime = 3600
	}

	indexer := &Indexer{
		db:               db,
		network:          network,
		chainID:          head.ChainID,
		indexName:        models.MempoolIndexName(network),
		filters:          indexerCfg.Filters,
		tzkt:             tzkt.NewTzKT(indexerCfg.DataSource.Tzkt.Struct().URL, indexerCfg.Filters.Addresses(), indexerCfg.Filters.Kinds),
		mempool:          memInd,
		prom:             prom,
		cache:            NewCache(2 * time.Hour),
		keepInChain:      uint64(delay) * settings.KeepInChainBlocks,
		keepOperations:   uint64(delay) * settings.ExpiredAfter,
		gasStatsLifetime: gasStatsLifetime,
		endorsements:     make(chan *models.Endorsement, 1024*32),
		rights:           ccache.New(ccache.Configure().MaxSize(60)),
		logger:           log.Logger.With().Str("network", network).Logger(),
		g:                workerpool.NewGroup(),
	}
	indexer.cache.Start(ctx)

	indexer.state = &database.State{
		IndexType: models.IndexTypeMempool,
		IndexName: indexer.indexName,
		Level:     head.Level,
	}

	for i := range indexerCfg.Filters.Kinds {
		if node.IsManager(indexerCfg.Filters.Kinds[i]) {
			indexer.hasManager = true
			break
		}
	}
	indexer.branches = newBlockQueue(expiredAfter, indexer.onPopBlockQueue, indexer.onRollbackBlockQueue)

	for _, kind := range indexer.filters.Kinds {
		if kind == node.KindEndorsement {
			indexer.delegates = newCachedDelegates(indexer.tzkt, constants.BlocksPerCycle)
			break
		}
	}

	return indexer, nil
}

// Start -
func (indexer *Indexer) Start(ctx context.Context) error {
	indexer.info().Strs("kinds", indexer.filters.Kinds).Msg("starting...")

	if err := indexer.initState(ctx); err != nil {
		return err
	}

	indexer.g.GoCtx(ctx, indexer.listen)

	if indexer.delegates != nil {
		if err := indexer.delegates.Init(ctx); err != nil {
			return err
		}

		indexer.g.GoCtx(ctx, indexer.setEndorsementBakers)

		var offset int
		for {
			endorsements, err := models.EndorsementsWithoutBaker(ctx, indexer.db.DB(), indexer.network, 100, offset)
			if err != nil {
				indexer.error(err).Msg("get endorsements without baker")
				break
			}
			for i := range endorsements {
				indexer.endorsements <- &endorsements[i]
			}

			if len(endorsements) < 100 {
				break
			}
			offset += len(endorsements)
		}
	}

	if err := indexer.tzkt.Connect(ctx); err != nil {
		return err
	}

	if err := indexer.tzkt.Subscribe(); err != nil {
		return err
	}

	indexer.mempool.Start(ctx)

	return nil
}

func (indexer *Indexer) sync(ctx context.Context) {
	indexer.info().Msg("start syncing...")
	indexer.g.GoCtx(ctx, func(ctx context.Context) {
		indexer.tzkt.Sync(ctx, indexer.state.Level)
	})
}

func (indexer *Indexer) initState(ctx context.Context) error {
	current, err := indexer.db.State(ctx, indexer.indexName)
	switch {
	case err == nil:
		indexer.state = current
	case errors.Is(err, sql.ErrNoRows):
		indexer.state = &database.State{
			IndexType: models.IndexTypeMempool,
			IndexName: indexer.indexName,
		}

		return indexer.db.CreateState(ctx, indexer.state)
	default:
		return err
	}

	return nil
}

// Close -
func (indexer *Indexer) Close() {
	indexer.g.Wait()
	indexer.info().Msg("indexer was stopped")
}

func (indexer *Indexer) close() error {
	indexer.info().Msg("stopping...")

	indexer.info().Msg("closing tzkt...")
	if err := indexer.tzkt.Close(); err != nil {
		return err
	}

	indexer.info().Msg("closing mempool...")
	if err := indexer.mempool.Close(); err != nil {
		return err
	}

	indexer.info().Msg("closing database...")
	if err := indexer.db.Close(); err != nil {
		return err
	}

	indexer.info().Msg("closing cache...")
	if err := indexer.cache.Close(); err != nil {
		return err
	}

	close(indexer.endorsements)

	return nil
}

func (indexer *Indexer) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			indexer.close()
			return
		case operations := <-indexer.tzkt.Operations():
			if err := indexer.handleInChain(ctx, operations); err != nil {
				indexer.error(err).Msg("handleInChain")
				continue
			}
		case block := <-indexer.tzkt.Blocks():
			if err := indexer.handleBlock(ctx, block); err != nil {
				indexer.error(err).Msg("handleBlock")
				continue
			}
		case msg := <-indexer.mempool.Operations():
			switch msg.Status {
			case receiver.StatusApplied:
				applied, ok := msg.Body.(node.Applied)
				if !ok {
					indexer.error(nil).Msgf("invalid applied operation %v", applied)
					continue
				}
				if !indexer.branches.Contains(applied.Branch) {
					continue
				}
				if indexer.isHashProcessed(applied.Hash) {
					continue
				}
				if err := indexer.handleAppliedOperation(ctx, applied, msg.Protocol); err != nil {
					indexer.error(err).Msg("handleAppliedOperation")
					continue
				}
			case receiver.StatusBranchDelayed, receiver.StatusBranchRefused, receiver.StatusRefused, receiver.StatusUnprocessed, receiver.StatusOutdated:
				failed, ok := msg.Body.(node.FailedMonitor)
				if !ok {
					indexer.error(nil).Msgf("invalid %s operation %v", msg.Status, failed)
					continue
				}

				if !indexer.branches.Contains(failed.Branch) {
					continue
				}
				if indexer.isHashProcessed(failed.Hash) {
					continue
				}
				if err := indexer.handleFailedOperation(ctx, failed, string(msg.Status), msg.Protocol); err != nil {
					indexer.error(err).Msg("handleFailedOperation")
					continue
				}
			default:
				indexer.error(nil).Msgf("invalid mempool operation status %s", msg.Status)
			}
		}
	}
}

func (indexer *Indexer) isHashProcessed(hash string) bool {
	key := fmt.Sprintf("hash:%s", hash)
	if !indexer.cache.Has(key) {
		indexer.cache.Set(key)
		return false
	}
	return true
}

func (indexer *Indexer) onPopBlockQueue(ctx context.Context, block Block) error {
	indexer.info().Uint64("block", block.Level).Msgf("operations with branch %s is expired", block.Branch)
	return models.SetExpired(ctx, indexer.db.DB(), indexer.network, block.Branch, indexer.filters.Kinds...)
}

func (indexer *Indexer) onRollbackBlockQueue(ctx context.Context, block Block) error {
	indexer.warn().Msgf("Rollback to %d level", block.Level)
	indexer.state.Level = block.Level
	indexer.state.Timestamp = block.Timestamp

	return indexer.db.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := models.Rollback(ctx, tx, indexer.network, block.Branch, block.Level, indexer.filters.Kinds...); err != nil {
			return err
		}
		_, err := tx.NewUpdate().Model(indexer.state).WherePK().Exec(ctx)
		return err
	})

}

func (indexer *Indexer) error(err error) *zerolog.Event {
	if err == nil {
		return indexer.logger.Error().Uint64("state", indexer.state.Level)
	}
	return indexer.logger.Err(err).Uint64("state", indexer.state.Level)
}

func (indexer *Indexer) info() *zerolog.Event {
	return indexer.logger.Info().Uint64("state", indexer.state.Level)
}

func (indexer *Indexer) warn() *zerolog.Event {
	return indexer.logger.Warn().Uint64("state", indexer.state.Level)
}
