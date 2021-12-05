package main

import (
	"context"

	"github.com/dipdup-net/mempool/cmd/mempool/endorsement"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
)

// CachedDelegates -
type CachedDelegates struct {
	Delegates map[string]PublicKey
	tzkt      *tzkt.TzKT

	blocksForCycle uint64
}

// PublicKey -
type PublicKey struct {
	Key    []byte
	Prefix string
}

func newCachedDelegates(tzkt *tzkt.TzKT, blocksForCycle uint64) *CachedDelegates {
	return &CachedDelegates{
		tzkt:           tzkt,
		Delegates:      make(map[string]PublicKey),
		blocksForCycle: blocksForCycle,
	}
}

// Update -
func (cd *CachedDelegates) Update(ctx context.Context, level uint64) error {
	if level%cd.blocksForCycle != 0 {
		return nil
	}
	return cd.Init(ctx)
}

// Init -
func (cd *CachedDelegates) Init(ctx context.Context) error {
	cd.Delegates = make(map[string]PublicKey)

	limit := int64(10000)

	var (
		offset int64
		end    bool
	)
	for !end {
		delegates, err := cd.tzkt.Delegates(ctx, limit, offset)
		if err != nil {
			return err
		}

		for i := range delegates {
			cd.Delegates[delegates[i].Address] = PublicKey{
				Prefix: delegates[i].PublicKey[:4],
				Key:    endorsement.DecodePublicKey(delegates[i].PublicKey),
			}
		}

		end = len(delegates) != int(limit)
	}

	return nil
}
