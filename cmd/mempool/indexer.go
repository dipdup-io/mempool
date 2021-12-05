package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pg "github.com/go-pg/pg/v10"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	generalConfig "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/receiver"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
)

// Indexer -
type Indexer struct {
	db               *pg.DB
	tzkt             *tzkt.TzKT
	mempool          *receiver.Receiver
	prom             *prometheus.Service
	branches         *BlockQueue
	cache            *Cache
	rights           *ccache.Cache
	delegates        *CachedDelegates
	state            models.State
	filters          config.Filters
	endorsements     chan *models.Endorsement
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
	constants, err := rpc.Constants(node.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if len(constants.TimeBetweenBlocks) == 0 {
		return nil, errors.Errorf("Empty time_between_blocks in node response: %s", network)
	}

	head, err := rpc.Header("head", node.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	memInd, err := receiver.New(indexerCfg.DataSource.RPC, network,
		receiver.WithStorage(db, constants.TimeBetweenBlocks[0]),
		receiver.WithPrometheus(prom),
	)
	if err != nil {
		return nil, err
	}

	expiredAfter := settings.ExpiredAfter
	if expiredAfter == 0 {
		metadata, err := rpc.HeadMetadata("head", node.WithContext(ctx))
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
		endorsements:     make(chan *models.Endorsement, 1024*32),
		rights:           ccache.New(ccache.Configure().MaxSize(60)),
	}

	indexer.state = models.State{
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
	indexer.log().WithField("kinds", indexer.filters.Kinds).Info("starting...")

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

		var offset int
		for {
			endorsements, err := models.EndorsementsWithoutBaker(indexer.db, indexer.network, 100, offset)
			if err != nil {
				log.Error(err)
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
	indexer.log().Info("start syncing...")
	indexer.wg.Add(1)
	go func() {
		defer indexer.wg.Done()
		indexer.tzkt.Sync(ctx, indexer.state.Level)
	}()

}

func (indexer *Indexer) initState() error {
	current, err := models.GetState(indexer.db, indexer.indexName)
	switch {
	case err == nil:
		indexer.state = current
	case errors.Is(err, pg.ErrNoRows):
		return nil
	default:
		return err
	}

	return nil
}

// Close -
func (indexer *Indexer) Close() {
	indexer.wg.Wait()
	indexer.log().Info("indexer was stopped")
}

func (indexer *Indexer) close() error {
	indexer.log().Info("stopping...")
	if err := indexer.tzkt.Close(); err != nil {
		return err
	}

	if err := indexer.mempool.Close(); err != nil {
		return err
	}
	if err := indexer.db.Close(); err != nil {
		return err
	}

	close(indexer.endorsements)

	return nil
}

func (indexer *Indexer) listen(ctx context.Context) {
	defer indexer.wg.Done()

	for {
		select {
		case <-ctx.Done():
			indexer.close()
			return
		case operations := <-indexer.tzkt.Operations():
			if err := indexer.handleInChain(ctx, operations); err != nil {
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
					indexer.log().Errorf("invalid applied operation %v", applied)
					continue
				}
				if indexer.isHashProcessed(applied.Hash) {
					continue
				}
				if err := indexer.handleAppliedOperation(ctx, applied, msg.Protocol); err != nil {
					log.Error(err)
					continue
				}
			case receiver.StatusBranchDelayed, receiver.StatusBranchRefused, receiver.StatusRefused, receiver.StatusUnprocessed:
				failed, ok := msg.Body.(node.FailedMonitor)
				if !ok {
					indexer.log().Errorf("invalid %s operation %v", msg.Status, failed)
					continue
				}
				if indexer.isHashProcessed(failed.Hash) {
					continue
				}
				if err := indexer.handleFailedOperation(ctx, failed, string(msg.Status), msg.Protocol); err != nil {
					indexer.log().Error(err)
					continue
				}
			default:
				indexer.log().Errorf("invalid mempool operation status %s", msg.Status)
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
	indexer.log().WithField("level", block.Level).Infof("operations with branch %s is expired", block.Branch)
	return models.SetExpired(indexer.db, indexer.network, block.Branch, indexer.filters.Kinds...)
}

func (indexer *Indexer) onRollbackBlockQueue(block Block) error {
	log.Warnf("Rollback to %d level", block.Level)

	tx, err := indexer.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	indexer.state.Level = block.Level
	if err := models.Rollback(tx, indexer.network, block.Branch, block.Level, indexer.filters.Kinds...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if err := indexer.state.Update(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (indexer *Indexer) log() *log.Entry {
	return log.WithField("state", indexer.state.Level).WithField("name", indexer.indexName)
}
