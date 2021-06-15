package main

import (
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
)

// Block -
type Block struct {
	Branch string
	Level  uint64
	Type   events.MessageType
}

func fromMessage(block tzkt.BlockMessage) Block {
	return Block{
		Branch: block.Hash,
		Level:  block.Level,
		Type:   block.Type,
	}
}

// BlockQueue -
type BlockQueue struct {
	queue      []Block
	levels     map[string]uint64
	onPop      func(block Block) error
	onRollback func(block Block) error
	capacity   uint64
}

func newBlockQueue(capacity uint64, onPop func(block Block) error, onRollback func(block Block) error) *BlockQueue {
	if capacity == 0 {
		capacity = 60
	}
	return &BlockQueue{
		queue:      make([]Block, 0, capacity),
		levels:     make(map[string]uint64),
		onPop:      onPop,
		onRollback: onRollback,
		capacity:   capacity,
	}
}

// Add -
func (bq *BlockQueue) Add(block tzkt.BlockMessage) error {
	b := fromMessage(block)

	switch block.Type {
	case events.MessageTypeState:
	case events.MessageTypeReorg:
		for item := bq.queue[len(bq.queue)-1]; len(bq.queue) > 0 && bq.queue[len(bq.queue)-1].Level > block.Level; item = bq.queue[len(bq.queue)-1] {
			if err := bq.onRollback(item); err != nil {
				return err
			}
			bq.queue = bq.queue[:len(bq.queue)-1]
			delete(bq.levels, item.Branch)
		}
	case events.MessageTypeData:
		if bq.Space() == 0 {
			item := bq.queue[0]
			bq.queue = bq.queue[1:]
			if bq.onPop != nil {
				if err := bq.onPop(item); err != nil {
					return err
				}
			}
			delete(bq.levels, item.Branch)
		}
		bq.queue = append(bq.queue, b)
		bq.levels[b.Branch] = b.Level + bq.capacity
	}

	return nil
}

// Space -
func (bq *BlockQueue) Space() uint64 {
	return bq.capacity - uint64(len(bq.queue))
}

// ExpirationLevel -
func (bq *BlockQueue) ExpirationLevel(hash string) uint64 {
	level, ok := bq.levels[hash]
	if ok {
		return level
	}
	return 0
}
