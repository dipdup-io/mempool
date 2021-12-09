package receiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	pg "github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Receiver -
type Receiver struct {
	urls      []string
	monitors  []*node.Monitor
	db        *database.PgGo
	prom      *prometheus.Service
	state     database.State
	indexName string
	protocol  string
	network   string

	blockTime int64

	wg         sync.WaitGroup
	operations chan Message
}

// New -
func New(urls []string, network string, opts ...ReceiverOption) (*Receiver, error) {
	if len(urls) == 0 {
		return nil, errors.Errorf("Empty url list: %s", network)
	}

	monitors := make([]*node.Monitor, len(urls))
	for i := range urls {
		monitors[i] = node.NewMonitor(urls[i])
	}

	indexer := Receiver{
		urls:       urls,
		operations: make(chan Message, 1024),
		indexName:  models.MempoolIndexName(network),
		network:    network,
		monitors:   monitors,
	}

	for i := range opts {
		opts[i](&indexer)
	}

	return &indexer, nil
}

// Start -
func (indexer *Receiver) Start(ctx context.Context) {
	if indexer.db != nil && len(indexer.urls) > 0 {
		indexer.wg.Add(1)
		go indexer.updateState(ctx, indexer.urls[0])
	}

	for i := range indexer.monitors {
		indexer.wg.Add(1)
		go indexer.run(ctx, indexer.monitors[i])

		indexer.monitors[i].SubscribeOnMempoolApplied(ctx)
		indexer.monitors[i].SubscribeOnMempoolBranchDelayed(ctx)
		indexer.monitors[i].SubscribeOnMempoolBranchRefused(ctx)
		indexer.monitors[i].SubscribeOnMempoolRefused(ctx)
	}
}

// Close -
func (indexer *Receiver) Close() error {
	indexer.wg.Wait()
	return nil
}

// Operations -
func (indexer *Receiver) Operations() <-chan Message {
	return indexer.operations
}

func (indexer *Receiver) run(ctx context.Context, monitor *node.Monitor) {
	defer indexer.wg.Done()

	for {
		select {
		case <-ctx.Done():
			for i := range indexer.monitors {
				if err := indexer.monitors[i].Close(); err != nil {
					log.Err(err).Msg("")
				}
			}

			close(indexer.operations)
			return
		case applied := <-monitor.Applied():
			for i := range applied {
				indexer.operations <- Message{
					Status:   StatusApplied,
					Body:     *applied[i],
					Protocol: indexer.protocol,
				}
			}
		case branchDelayed := <-monitor.BranchDelayed():
			for i := range branchDelayed {
				indexer.operations <- Message{
					Status:   StatusBranchDelayed,
					Body:     *branchDelayed[i],
					Protocol: indexer.protocol,
				}
			}
		case branchRefused := <-monitor.BranchRefused():
			for i := range branchRefused {
				indexer.operations <- Message{
					Status:   StatusBranchRefused,
					Body:     *branchRefused[i],
					Protocol: indexer.protocol,
				}
			}
		case refused := <-monitor.Refused():
			for i := range refused {
				indexer.operations <- Message{
					Status:   StatusRefused,
					Body:     *refused[i],
					Protocol: indexer.protocol,
				}
			}
		}
	}
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

func (indexer *Receiver) updateState(ctx context.Context, url string) {
	defer indexer.wg.Done()

	rpc := node.NewNodeRPC(url)

	ticker := time.NewTicker(time.Second * time.Duration(indexer.blockTime))
	defer ticker.Stop()

	// init
	if err := indexer.checkHead(ctx, rpc); err != nil {
		log.Err(err).Msg("")
	}
	if err := indexer.setState(); err != nil {
		log.Err(err).Msg("")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := indexer.checkHead(ctx, rpc); err != nil {
				log.Err(err).Msg("")
				continue
			}
			if err := indexer.setState(); err != nil {
				log.Err(err).Msg("")
				continue
			}
		}
	}
}

func (indexer *Receiver) setState() error {
	state, err := indexer.db.State(indexer.indexName)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
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
