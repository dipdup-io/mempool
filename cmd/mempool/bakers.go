package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tools/forge"
	"github.com/dipdup-net/go-lib/tzkt/data"
	"github.com/dipdup-net/mempool/cmd/mempool/endorsement"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/rs/zerolog/log"
)

const unknownBaker = "unknown"

func (indexer *Indexer) setEndorsementBakers(ctx context.Context) {
	defer indexer.wg.Done()

	indexer.info().Msg("Thread for finding endorsement baker started")

	for {
		select {
		case <-ctx.Done():
			return
		case endorsement := <-indexer.endorsements:
			if len(endorsement.Errors) > 0 {
				if err := indexer.findBaker(ctx, endorsement); err != nil {
					log.Err(err).Msg("find baker")
					continue
				}
			} else {
				endorsement.Baker = unknownBaker
			}
			if _, err := indexer.db.DB().Model(endorsement).WherePK().Update("baker", endorsement.Baker); err != nil {
				log.Err(err).Msg("set baker to endorsement")
			}
		}
	}
}

func (indexer *Indexer) getEndorsingRights(ctx context.Context, level uint64) ([]data.Right, error) {
	rights, err := indexer.rights.Fetch(fmt.Sprintf("rights/%s/%d", indexer.network, level), 15*time.Minute, func() (interface{}, error) {
		rights, err := indexer.tzkt.Rights(ctx, level)
		if err != nil {
			return nil, err
		}

		sort.Sort(BySlots(rights))
		return rights, nil
	})
	if err != nil {
		return nil, err
	}
	if result, ok := rights.Value().([]data.Right); !ok {
		return nil, errors.New("invalid rights type")
	} else {
		return result, nil
	}
}

func (indexer *Indexer) findBaker(ctx context.Context, e *models.Endorsement) error {
	if err := indexer.delegates.Update(ctx, e.Level); err != nil {
		return err
	}

	rights, err := indexer.getEndorsingRights(ctx, e.Level)
	if err != nil {
		return err
	}

	forged, err := forge.Endorsement(node.Endorsement{
		Level:    e.Level,
		Metadata: &node.EndorsementMetadata{},
	}, e.Branch)
	if err != nil {
		return err
	}

	hash := endorsement.Hash(indexer.chainID, forged)
	decodedSignature := endorsement.DecodeSignature(e.Signature)

	for i := len(rights) - 1; i >= 0; i-- {
		if rights[i].Slots == 0 {
			break
		}
		address := rights[i].Baker.Address
		publicKey, ok := indexer.delegates.Delegates[address]
		if !ok {
			continue
		}
		if !endorsement.CheckKey(publicKey.Prefix, publicKey.Key, decodedSignature, hash) {
			continue
		}
		e.Baker = address
		break
	}
	if e.Baker == "" {
		e.Baker = "unknown"
	}

	return nil
}

// BySlots -
type BySlots []data.Right

// Len -
func (rights BySlots) Len() int { return len(rights) }

// Less -
func (rights BySlots) Less(i, j int) bool { return rights[i].Slots < rights[j].Slots }

// Swap -
func (rights BySlots) Swap(i, j int) { rights[i], rights[j] = rights[j], rights[i] }
