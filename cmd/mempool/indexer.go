package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	generalConfig "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/state"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/receiver"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	"gorm.io/gorm"
)

// Indexer -
type Indexer struct {
	db           *gorm.DB
	tzkt         *tzkt.TzKT
	mempool      *receiver.Receiver
	manager      *Manager
	state        state.State
	network      string
	indexName    string
	chainID      string
	filters      config.Filters
	branches     *BlockQueue
	cache        *ccache.Cache
	delegates    *CachedDelegates
	threadsCount int

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewIndexer -
func NewIndexer(network string, indexerCfg config.Indexer, database generalConfig.Database, settings config.Settings) (*Indexer, error) {
	db, err := models.OpenDatabaseConnection(database, indexerCfg.Filters.Kinds...)
	if err != nil {
		return nil, err
	}

	rpc := node.NewNodeRPC(indexerCfg.DataSource.RPC[0], node.WithTimeout(settings.RPCTimeout))
	constants, err := rpc.Constants()
	if err != nil {
		return nil, err
	}
	if len(constants.TimeBetweenBlocks) == 0 {
		return nil, errors.Errorf("Empty time_between_blocks in node response: %s", network)
	}

	head, err := rpc.Head()
	if err != nil {
		return nil, err
	}

	memInd, err := receiver.New(indexerCfg.DataSource.RPC, network,
		receiver.WithInterval(settings.MempoolRequestInterval),
		receiver.WithTimeout(settings.RPCTimeout),
		receiver.WithStorage(db, constants.TimeBetweenBlocks[0]),
	)
	if err != nil {
		return nil, err
	}

	indexer := &Indexer{
		db:           db,
		network:      network,
		chainID:      head.ChainID,
		indexName:    models.MempoolIndexName(network),
		filters:      indexerCfg.Filters,
		tzkt:         tzkt.NewTzKT(indexerCfg.DataSource.Tzkt, indexerCfg.Filters.Accounts, indexerCfg.Filters.Kinds),
		mempool:      memInd,
		manager:      NewManager(db, settings, uint64(constants.TimeBetweenBlocks[0]), indexerCfg.Filters.Kinds...),
		cache:        ccache.New(ccache.Configure().MaxSize(2 ^ 13)),
		threadsCount: 1,
	}

	indexer.branches = newBlockQueue(settings.ExpiredAfter, indexer.onPopBlockQueue, indexer.onRollbackBlockQueue)

	for _, kind := range indexer.filters.Kinds {
		if kind == node.KindEndorsement {
			indexer.delegates = newCachedDelegates(indexer.tzkt, constants.BlocksPerCycle)
			indexer.threadsCount += 1
			break
		}
	}

	indexer.stop = make(chan struct{}, indexer.threadsCount)

	return indexer, nil
}

// Start -
func (indexer *Indexer) Start() error {
	indexer.log().WithField("kinds", indexer.filters.Kinds).Info("Starting...")

	if err := indexer.initState(); err != nil {
		return err
	}

	indexer.wg.Add(1)
	go indexer.listen()

	if indexer.delegates != nil {
		if err := indexer.delegates.Init(); err != nil {
			return err
		}
		indexer.wg.Add(1)
		go indexer.setEndorsementBakers()
	}

	go indexer.sync()

	go indexer.mempool.Start()
	indexer.manager.Start()

	return nil
}

func (indexer *Indexer) sync() {
	indexer.log().Info("Start syncing...")
	indexer.tzkt.Sync(indexer.state.Level)
}

func (indexer *Indexer) initState() error {
	current, err := state.Get(indexer.db, indexer.indexName)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		indexer.state = state.State{
			IndexType: models.IndexTypeMempool,
			IndexName: indexer.indexName,
		}
	} else {
		indexer.state = current
	}

	return nil
}

// Close -
func (indexer *Indexer) Close() error {
	indexer.log().Info("Stopping...")
	for i := 0; i < 2; i++ {
		indexer.stop <- struct{}{}
	}
	indexer.wg.Wait()

	if err := indexer.manager.Close(); err != nil {
		return err
	}

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

	close(indexer.stop)

	indexer.log().Info("Indexer was stopped")
	return nil
}

func (indexer *Indexer) listen() {
	defer indexer.wg.Done()

	for {
		select {
		case <-indexer.stop:
			return
		case operations := <-indexer.tzkt.Operations():
			if err := indexer.handleInChain(operations); err != nil {
				indexer.log().Error(err)
				continue
			}
		case block := <-indexer.tzkt.Blocks():
			if err := indexer.handleBlock(block); err != nil {
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
				if err := indexer.handleAppliedOperation(applied); err != nil {
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
				if err := indexer.handleFailedOperation(failed, string(msg.Status)); err != nil {
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
	item := indexer.cache.Get(key)
	if item == nil {
		indexer.cache.Set(key, struct{}{}, time.Minute*10)
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
