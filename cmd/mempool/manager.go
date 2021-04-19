package main

import (
	"sync"
	"time"

	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager -
type Manager struct {
	db *gorm.DB

	keepOperations   uint64
	keepInChain      uint64
	lostAfterSeconds uint64
	blockTime        uint64
	kinds            []string

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewManager -
func NewManager(db *gorm.DB, settings config.Settings, blockTime uint64, kinds ...string) *Manager {
	return &Manager{
		db:               db,
		keepOperations:   settings.KeepOperations,
		keepInChain:      blockTime * settings.KeepInChainBlocks,
		lostAfterSeconds: blockTime * settings.ExpiredAfter,
		blockTime:        blockTime,
		stop:             make(chan struct{}, 1),
		kinds:            kinds,
	}
}

// Start -
func (manager *Manager) Start() {
	manager.wg.Add(1)
	go manager.work()
}

// Close -
func (manager *Manager) Close() error {
	manager.stop <- struct{}{}
	manager.wg.Wait()

	close(manager.stop)
	return nil
}

func (manager *Manager) work() {
	defer manager.wg.Done()

	blockTicker := time.NewTicker(time.Second * time.Duration(manager.blockTime))
	defer blockTicker.Stop()

	for {
		select {
		case <-manager.stop:
			return
		case <-blockTicker.C:
			err := manager.db.Transaction(func(tx *gorm.DB) error {
				if err := models.DeleteOldOperations(tx, manager.keepInChain, models.StatusInChain, manager.kinds...); err != nil {
					return errors.Wrap(err, "DeleteOldOperations in_chain")
				}
				if err := models.DeleteOldOperations(tx, manager.keepOperations, "", manager.kinds...); err != nil {
					return errors.Wrap(err, "DeleteOldOperations")
				}
				return nil
			})

			if err != nil {
				log.Error(err)
				continue
			}

		}
	}
}
