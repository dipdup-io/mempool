package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	generalConfig "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/go-lib/state"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/receiver"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	"gorm.io/gorm"
)

// Indexer -
type Indexer struct {
	db               *gorm.DB
	tzkt             *tzkt.TzKT
	mempool          *receiver.Receiver
	prom             *prometheus.Service
	branches         *BlockQueue
	cache            *Cache
	delegates        *CachedDelegates
	state            state.State
	filters          config.Filters
	network          string
	indexName        string
	chainID          string
	keepInChain      uint64
	keepOperations   uint64
	gasStatsLifetime uint64
	hasManager       bool

	wg sync.WaitGroup
}

// NewIndexer -
func NewIndexer(ctx context.Context, network string, indexerCfg config.Indexer, database generalConfig.Database, settings config.Settings, prom *prometheus.Service) (*Indexer, error) {
	db, err := models.OpenDatabaseConnection(ctx, database, indexerCfg.Filters.Kinds...)
	if err != nil {
		return nil, err
	}

	rpc := node.NewNodeRPC(indexerCfg.DataSource.RPC[0])
	constants, err := rpc.Constants()
	if err != nil {
		return nil, err
	}
	if len(constants.TimeBetweenBlocks) == 0 {
		return nil, errors.Errorf("Empty time_between_blocks in node response: %s", network)
	}

	head, err := rpc.Header("head")
	if err != nil {
		return nil, err
	}

	memInd, err := receiver.New(indexerCfg.DataSource.RPC, network,
		receiver.WithInterval(settings.MempoolRequestInterval),
		receiver.WithTimeout(settings.RPCTimeout),
		receiver.WithStorage(db, constants.TimeBetweenBlocks[0]),
		receiver.WithPrometheus(prom),
	)
	if err != nil {
		return nil, err
	}

	expiredAfter := settings.ExpiredAfter
	if expiredAfter == 0 {
		metadata, err := rpc.HeadMetadata("head")
		if err != nil {
			return nil, err
		}
		expiredAfter = metadata.MaxOperationsTTL
	}

	indexer := &Indexer{
		db:               db,
		network:          network,
		chainID:          head.ChainID,
		indexName:        models.MempoolIndexName(network),
		filters:          indexerCfg.Filters,
		tzkt:             tzkt.NewTzKT(indexerCfg.DataSource.Tzkt, indexerCfg.Filters.Accounts, indexerCfg.Filters.Kinds),
		mempool:          memInd,
		prom:             prom,
		cache:            NewCache(2 * time.Hour),
		keepInChain:      uint64(constants.TimeBetweenBlocks[0]) * settings.KeepInChainBlocks,
		keepOperations:   uint64(constants.TimeBetweenBlocks[0]) * settings.ExpiredAfter,
		gasStatsLifetime: settings.GasStatsLifetime,
	}

	indexer.state = state.State{
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
	indexer.log().WithField("kinds", indexer.filters.Kinds).Info("Starting...")

	if err := indexer.initState(); err != nil {
		return err
	}

	indexer.wg.Add(1)
	go indexer.listen(ctx)

	if indexer.delegates != nil {
		if err := indexer.delegates.Init(ctx); err != nil {
			return err
		}
		indexer.wg.Add(1)
		go indexer.setEndorsementBakers(ctx)
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
	indexer.wg.Add(1)
	indexer.log().Info("start syncing...")
	go func() {
		defer indexer.wg.Done()
		indexer.tzkt.Sync(ctx, indexer.state.Level)
	}()

}

func (indexer *Indexer) initState() error {
	current, err := state.Get(indexer.db, indexer.indexName)
	switch {
	case err == nil:
		indexer.state = current
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	default:
		return err
	}

	return nil
}

// Close -
func (indexer *Indexer) Close() error {
	indexer.log().Info("Stopping...")
	indexer.wg.Wait()

	if err := indexer.tzkt.Close(); err != nil {
		return err
	}

	if err := indexer.mempool.Close(); err != nil {
		return err
	}
	sqlDB, err := indexer.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}

	indexer.log().Info("Indexer was stopped")
	return nil
}

func (indexer *Indexer) listen(ctx context.Context) {
	defer indexer.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case operations := <-indexer.tzkt.Operations():
			if err := indexer.handleInChain(operations); err != nil {
				indexer.log().Error(err)
				continue
			}
		case block := <-indexer.tzkt.Blocks():
			if err := indexer.handleBlock(ctx, block); err != nil {
				indexer.log().Error(err)
				continue
			}
		case msg := <-indexer.mempool.Operations():
			switch msg.Status {
			case receiver.StatusApplied:
				applied, ok := msg.Body.(node.Applied)
				if !ok {
					indexer.log().Errorf("Invalid applied operation %v", applied)
					continue
				}
				if indexer.isHashProcessed(applied.Hash) {
					continue
				}
				if err := indexer.handleAppliedOperation(applied, msg.Protocol); err != nil {
					log.Error(err)
					continue
				}
			case receiver.StatusBranchDelayed, receiver.StatusBranchRefused, receiver.StatusRefused, receiver.StatusUnprocessed:
				failed, ok := msg.Body.(node.Failed)
				if !ok {
					indexer.log().Errorf("Invalid %s operation %v", msg.Status, failed)
					continue
				}
				if indexer.isHashProcessed(failed.Hash) {
					continue
				}
				if err := indexer.handleFailedOperation(failed, string(msg.Status), msg.Protocol); err != nil {
					indexer.log().Error(err)
					continue
				}
			default:
				indexer.log().Errorf("Invalid mempool operation status %s", msg.Status)
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

func (indexer *Indexer) onPopBlockQueue(block Block) error {
	indexer.log().WithField("level", block.Level).Infof("Operations with branch %s is expired", block.Branch)
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		return models.SetExpired(tx, indexer.network, block.Branch, indexer.filters.Kinds...)
	})
}

func (indexer *Indexer) onRollbackBlockQueue(block Block) error {
	log.Warnf("Rollback to %d level", block.Level)
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		if err := models.Rollback(tx, indexer.network, block.Branch, block.Level, indexer.filters.Kinds...); err != nil {
			return err
		}
		indexer.state.Level = block.Level
		return indexer.state.Update(tx)
	})
}

func (indexer *Indexer) log() *log.Entry {
	return log.WithField("state", indexer.state.Level).WithField("name", indexer.indexName)
}
