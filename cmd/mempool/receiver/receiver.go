package receiver

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dipdup-io/workerpool"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/prometheus"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Receiver -
type Receiver struct {
	url       string
	monitor   *node.Monitor
	db        *database.Bun
	prom      *prometheus.Service
	state     *database.State
	indexName string
	protocol  string
	network   string

	blockTime int64

	g          workerpool.Group
	operations chan Message
}

// New -
func New(url string, network string, db *database.Bun, opts ...ReceiverOption) (*Receiver, error) {
	if url == "" {
		return nil, errors.Errorf("empty url: %s", network)
	}
	if db == nil {
		return nil, errors.Errorf("nil database connection: %s", network)
	}

	indexer := Receiver{
		url:        url,
		db:         db,
		operations: make(chan Message, 1024),
		indexName:  models.MempoolIndexName(network),
		network:    network,
		monitor:    node.NewMonitor(url),
		g:          workerpool.NewGroup(),
	}

	for i := range opts {
		opts[i](&indexer)
	}

	if indexer.blockTime == 0 {
		indexer.blockTime = 15
	}

	return &indexer, nil
}

// Start -
func (indexer *Receiver) Start(ctx context.Context) {
	indexer.g.GoCtx(ctx, func(ctx context.Context) {
		indexer.updateState(ctx, indexer.url)
	})

	indexer.g.GoCtx(ctx, func(ctx context.Context) {
		indexer.run(ctx, indexer.monitor)
	})

	indexer.monitor.SubscribeOnMempoolApplied(ctx)
	indexer.monitor.SubscribeOnMempoolBranchDelayed(ctx)
	indexer.monitor.SubscribeOnMempoolBranchRefused(ctx)
	indexer.monitor.SubscribeOnMempoolRefused(ctx)
	indexer.monitor.SubscribeOnMempoolOutdated(ctx)
}

// Close -
func (indexer *Receiver) Close() error {
	indexer.g.Wait()

	if err := indexer.monitor.Close(); err != nil {
		log.Err(err).Str("network", indexer.network).Msg("closing monitor")
	}

	close(indexer.operations)
	return nil
}

// Operations -
func (indexer *Receiver) Operations() <-chan Message {
	return indexer.operations
}

func (indexer *Receiver) run(ctx context.Context, monitor *node.Monitor) {
	for {
		select {
		case <-ctx.Done():
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
		case outdated := <-monitor.Outdated():
			for i := range outdated {
				indexer.operations <- Message{
					Status:   StatusOutdated,
					Body:     *outdated[i],
					Protocol: indexer.protocol,
				}
			}
		}
	}
}

func (indexer *Receiver) checkHead(ctx context.Context, rpc node.API) error {
	head, err := rpc.Header(ctx, "head")
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
	ticker := time.NewTicker(time.Second * time.Duration(indexer.blockTime))
	defer ticker.Stop()

	// init
	if err := indexer.setState(ctx); err != nil {
		log.Err(err).Msg("set state")
	}

	rpc := node.NewMainRPC(url)
	if err := indexer.checkHead(ctx, rpc); err != nil {
		log.Err(err).Msg("check head")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := indexer.checkHead(ctx, rpc); err != nil {
				log.Err(err).Msg("check head")
				continue
			}
			if err := indexer.setState(ctx); err != nil {
				log.Err(err).Msg("set state")
				continue
			}
		}
	}
}

func (indexer *Receiver) setState(ctx context.Context) error {
	state, err := indexer.db.State(ctx, indexer.indexName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			indexer.state = new(database.State)
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
