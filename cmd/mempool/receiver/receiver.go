package receiver

import (
	"sync"
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/state"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Receiver -
type Receiver struct {
	urls      []string
	db        *gorm.DB
	state     state.State
	indexName string
	protocol  string

	blockTime int64
	interval  uint64
	timeout   uint64

	wg         sync.WaitGroup
	stop       chan struct{}
	operations chan Message
}

// New -
func New(urls []string, network string, opts ...ReceiverOption) (*Receiver, error) {
	if len(urls) == 0 {
		return nil, errors.Errorf("Empty url list: %s", network)
	}
	indexer := Receiver{
		urls:       urls,
		stop:       make(chan struct{}, len(urls)+1),
		operations: make(chan Message, 1024),
		indexName:  models.MempoolIndexName(network),
	}

	for i := range opts {
		opts[i](&indexer)
	}

	return &indexer, nil
}

// Start -
func (indexer *Receiver) Start() {
	if indexer.db != nil {
		indexer.wg.Add(1)
		go indexer.updateState()
	}

	for i := range indexer.urls {
		indexer.wg.Add(1)
		go indexer.run(indexer.urls[i])
	}
}

// Close -
func (indexer *Receiver) Close() error {
	for range indexer.urls {
		indexer.stop <- struct{}{}
	}
	if indexer.db != nil {
		indexer.stop <- struct{}{}
	}
	indexer.wg.Wait()

	close(indexer.operations)
	close(indexer.stop)
	return nil
}

// Operations -
func (indexer *Receiver) Operations() <-chan Message {
	return indexer.operations
}

func (indexer *Receiver) run(url string) {
	defer indexer.wg.Done()

	rpc := node.NewNodeRPC(url, node.WithTimeout(indexer.timeout))

	for {
		select {
		case <-indexer.stop:
			return
		default:
			if err := indexer.process(rpc); err != nil {
				log.Error(err)
			}
			time.Sleep(time.Duration(indexer.interval) * time.Second)
		}
	}
}

func (indexer *Receiver) process(rpc *node.NodeRPC) error {
	if err := indexer.checkHead(rpc); err != nil {
		return err
	}

	result, err := rpc.PendingOperations()
	if err != nil {
		return err
	}
	for _, operation := range result.Applied {
		indexer.operations <- Message{
			Status:   StatusApplied,
			Body:     operation,
			Protocol: indexer.protocol,
		}
	}
	for _, operation := range result.BranchDelayed {
		indexer.operations <- Message{
			Status:   StatusBranchDelayed,
			Body:     operation,
			Protocol: indexer.protocol,
		}
	}
	for _, operation := range result.BranchRefused {
		indexer.operations <- Message{
			Status:   StatusBranchRefused,
			Body:     operation,
			Protocol: indexer.protocol,
		}
	}
	for _, operation := range result.Refused {
		indexer.operations <- Message{
			Status:   StatusRefused,
			Body:     operation,
			Protocol: indexer.protocol,
		}
	}
	return nil
}

func (indexer *Receiver) checkHead(rpc *node.NodeRPC) error {
	if indexer.db == nil {
		return nil
	}

	head, err := rpc.Header()
	if err != nil {
		return err
	}

	// If node is behind indexer more than one block throw error
	if head.Level < indexer.state.Level-1 && indexer.state.Level > 0 {
		return errors.Errorf("Node is stucked url=%s node_level=%d indexer_level=%d", rpc.URL(), head.Level, indexer.state.Level)
	}

	indexer.protocol = head.Protocol
	return nil
}

func (indexer *Receiver) updateState() {
	defer indexer.wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(indexer.blockTime))
	defer ticker.Stop()

	// init
	if err := indexer.setState(); err != nil {
		log.Error(err)
	}

	for {
		select {
		case <-indexer.stop:
			return
		case <-ticker.C:
			if err := indexer.setState(); err != nil {
				log.Error(err)
				continue
			}
		}
	}
}

func (indexer *Receiver) setState() error {
	state, err := state.Get(indexer.db, indexer.indexName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	indexer.state = state
	return nil
}
