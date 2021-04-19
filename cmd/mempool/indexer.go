package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/receiver"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	"github.com/dipdup-net/mempool/internal/config"
	"github.com/dipdup-net/mempool/internal/node"
	"github.com/dipdup-net/mempool/internal/state"
	"gorm.io/gorm"
)

// Indexer -
type Indexer struct {
	db              *gorm.DB
	externalIndexer *tzkt.TzKT
	mempool         *receiver.Receiver
	manager         *Manager
	state           state.State
	network         string
	indexName       string
	filters         config.Filters
	branches        *BlockQueue
	cache           *ccache.Cache

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewIndexer -
func NewIndexer(network string, indexerCfg config.Indexer, database config.Database, settings config.Settings) (*Indexer, error) {
	db, err := models.OpenDatabaseConnection(database, indexerCfg.Filters.Kinds...)
	if err != nil {
		return nil, err
	}
	constants, err := node.NewNodeRPC(indexerCfg.DataSource.RPC[0], node.WithTimeout(settings.RPCTimeout)).Constants()
	if err != nil {
		return nil, err
	}
	if len(constants.TimeBetweenBlocks) == 0 {
		return nil, errors.Errorf("Empty time_between_blocks in node response: %s", network)
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
		db:              db,
		network:         network,
		indexName:       models.MempoolIndexName(network),
		filters:         indexerCfg.Filters,
		externalIndexer: tzkt.NewTzKT(indexerCfg.DataSource.Tzkt, indexerCfg.Filters.Kinds),
		mempool:         memInd,
		manager:         NewManager(db, settings, uint64(constants.TimeBetweenBlocks[0]), indexerCfg.Filters.Kinds...),
		cache:           ccache.New(ccache.Configure().MaxSize(2 ^ 13)),
		stop:            make(chan struct{}, 1),
	}

	indexer.branches = newBlockQueue(settings.ExpiredAfter, indexer.onPopBlockQueue, indexer.onRollbackBlockQueue)

	return indexer, nil
}

// Start -
func (indexer *Indexer) Start() error {
	if err := indexer.initState(); err != nil {
		return err
	}

	indexer.wg.Add(1)
	go indexer.listen()

	if indexer.state.Level > 0 {
		indexer.log().Info("Start syncing...")
		if err := indexer.externalIndexer.Sync(indexer.state.Level); err != nil {
			return err
		}
	}

	if err := indexer.externalIndexer.Connect(); err != nil {
		return err
	}

	if err := indexer.subscribe(); err != nil {
		return err
	}

	indexer.log().Info("Start indexing...")
	indexer.mempool.Start()

	indexer.manager.Start()

	return nil
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

func (indexer *Indexer) subscribe() error {
	if err := indexer.externalIndexer.SubscribeToBlocks(); err != nil {
		return err
	}

	if len(indexer.filters.Accounts) == 0 {
		return indexer.externalIndexer.SubscribeToOperations("", indexer.filters.Kinds...)
	}

	for _, account := range indexer.filters.Accounts {
		if err := indexer.externalIndexer.SubscribeToOperations(account, indexer.filters.Kinds...); err != nil {
			return err
		}
	}
	return nil
}

// Close -
func (indexer *Indexer) Close() error {
	indexer.stop <- struct{}{}
	indexer.wg.Wait()

	if err := indexer.manager.Close(); err != nil {
		return err
	}

	if err := indexer.externalIndexer.Close(); err != nil {
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
		case operations := <-indexer.externalIndexer.Operations():
			if err := indexer.handleInChain(operations); err != nil {
				indexer.log().Error(err)
				continue
			}
		case block := <-indexer.externalIndexer.Blocks():
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
