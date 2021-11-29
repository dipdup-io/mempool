package main

import (
	"context"
	"time"

	"github.com/dipdup-net/go-lib/tzkt/api"
	"github.com/dipdup-net/mempool/cmd/mempool/endorsement"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (indexer *Indexer) setEndorsementBakers(ctx context.Context) {
	defer indexer.wg.Done()

	log.WithField("network", indexer.network).Info("Thread for finding endorsement baker started")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	var currentLevel uint64
	var rights []api.Right

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := indexer.db.Transaction(func(tx *gorm.DB) error {
				endorsements, err := models.EndorsementsWithoutBaker(tx)
				if err != nil {
					return err
				}

				for _, e := range endorsements {
					if err := indexer.delegates.Update(ctx, e.Level); err != nil {
						return err
					}

					if currentLevel != e.Level {
						rights, err = indexer.tzkt.Rights(ctx, e.Level+1)
						if err != nil {
							return err
						}
						currentLevel = e.Level
					}

					forged, err := e.Forge()
					if err != nil {
						return err
					}

					for i := range rights {
						address := rights[i].Baker.Address
						publicKey, ok := indexer.delegates.Delegates[address]
						if !ok {
							continue
						}
						if !endorsement.CheckKey(publicKey, e.Signature, indexer.chainID, forged) {
							continue
						}
						if err := tx.Model(&e).Update("baker", address).Error; err != nil {
							return err
						}
						break
					}
				}

				return nil
			}); err != nil {
				log.Error(err)
				continue
			}
		}
	}
}
