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
	onPop      func(block Block) error
	onRollback func(block Block) error
}

func newBlockQueue(capacity uint64, onPop func(block Block) error, onRollback func(block Block) error) *BlockQueue {
	if capacity == 0 {
		capacity = 60
	}
	return &BlockQueue{
		queue:      make([]Block, 0, capacity),
		onPop:      onPop,
		onRollback: onRollback,
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
		}
		bq.queue = append(bq.queue, b)
	}

	return nil
}

// Space -
func (bq *BlockQueue) Space() uint64 {
	return uint64(cap(bq.queue) - len(bq.queue))
}
