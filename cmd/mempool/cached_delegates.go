package main

import (
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
)

// CachedDelegates -
type CachedDelegates struct {
	Delegates map[string]string
	tzkt      *tzkt.TzKT

	blocksForCycle uint64
}

func newCachedDelegates(tzkt *tzkt.TzKT, blocksForCycle uint64) *CachedDelegates {
	return &CachedDelegates{
		tzkt:           tzkt,
		Delegates:      make(map[string]string),
		blocksForCycle: blocksForCycle,
	}
}

// Update -
func (cd *CachedDelegates) Update(level uint64) error {
	if level%cd.blocksForCycle != 0 {
		return nil
	}
	return cd.Init()
}

// Init -
func (cd *CachedDelegates) Init() error {
	delegates, err := cd.tzkt.Delegates()
	if err != nil {
		return err
	}

	cd.Delegates = make(map[string]string)
	for i := range delegates {
		cd.Delegates[delegates[i].Address] = delegates[i].PublicKey
	}

	return nil
}
