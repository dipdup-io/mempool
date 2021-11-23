package receiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
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
	prom      *prometheus.Service
	state     state.State
	indexName string
	protocol  string
	network   string

	blockTime int64
	interval  uint64
	timeout   uint64

	wg         sync.WaitGroup
	operations chan Message
}

// New -
func New(urls []string, network string, opts ...ReceiverOption) (*Receiver, error) {
	if len(urls) == 0 {
		return nil, errors.Errorf("Empty url list: %s", network)
	}
	indexer := Receiver{
		urls:       urls,
		operations: make(chan Message, 1024),
		indexName:  models.MempoolIndexName(network),
		network:    network,
	}

	for i := range opts {
		opts[i](&indexer)
	}

	return &indexer, nil
}

// Start -
func (indexer *Receiver) Start(ctx context.Context) {
	if indexer.db != nil {
		indexer.wg.Add(1)
		go indexer.updateState(ctx)
	}

	for i := range indexer.urls {
		indexer.wg.Add(1)
		go indexer.run(ctx, indexer.urls[i])
	}
}

// Close -
func (indexer *Receiver) Close() error {
	indexer.wg.Wait()

	close(indexer.operations)
	return nil
}

// Operations -
func (indexer *Receiver) Operations() <-chan Message {
	return indexer.operations
}

func (indexer *Receiver) run(ctx context.Context, url string) {
	defer indexer.wg.Done()

	rpc := node.NewNodeRPC(url)
	if err := indexer.process(ctx, rpc); err != nil {
		log.Error(err)
	}

	ticker := time.NewTicker(time.Duration(indexer.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := indexer.process(ctx, rpc); err != nil {
				log.Error(err)
			}
		}
	}
}

func (indexer *Receiver) process(ctx context.Context, rpc *node.NodeRPC) error {
	if err := indexer.checkHead(ctx, rpc); err != nil {
		return err
	}

	result, err := rpc.PendingOperations(node.WithContext(ctx))
	if err != nil {
		indexer.incrementMetric(rpc.URL(), indexer.network, err)
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

func (indexer *Receiver) checkHead(ctx context.Context, rpc *node.NodeRPC) error {
	if indexer.db == nil {
		return nil
	}

	head, err := rpc.Header("head", node.WithContext(ctx))
	if err != nil {
		indexer.incrementMetric(rpc.URL(), indexer.network, err)
		return err
	}

	// If node is behind indexer more than one block throw error
	if head.Level < indexer.state.Level-1 && indexer.state.Level > 0 {
		return errors.Errorf("Node is stucked url=%s node_level=%d indexer_level=%d", rpc.URL(), head.Level, indexer.state.Level)
	}

	indexer.protocol = head.Protocol
	return nil
}

func (indexer *Receiver) updateState(ctx context.Context) {
	defer indexer.wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(indexer.blockTime))
	defer ticker.Stop()

	// init
	if err := indexer.setState(); err != nil {
		log.Error(err)
	}

	for {
		select {
		case <-ctx.Done():
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

func (indexer *Receiver) incrementMetric(url, network string, err error) {
	if err == nil || indexer.prom == nil {
		return
	}

	reqErr, ok := err.(node.RequestError)
	if !ok {
		return
	}
	indexer.prom.IncrementCounter("mempool_rpc_errors_count", map[string]string{
		"network": network,
		"node":    url,
		"code":    fmt.Sprintf("%d", reqErr.Code),
	})
}
