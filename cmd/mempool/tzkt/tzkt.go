package tzkt

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dipdup-net/go-lib/tzkt/api"
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	pageSize = 1000
)

// TzKT - tzkt data source
type TzKT struct {
	api      *api.API
	client   *events.TzKT
	state    uint64
	kinds    []string
	accounts []string

	operations chan OperationMessage
	blocks     chan BlockMessage
	wg         sync.WaitGroup
}

// NewTzKT - TzKT constructor
func NewTzKT(url string, accounts []string, kinds []string) *TzKT {
	tzktKinds := make([]string, 0)
	for i := range kinds {
		if kind, ok := toTzKTKinds[kinds[i]]; ok {
			tzktKinds = append(tzktKinds, kind)
		}
	}
	return &TzKT{
		client:     events.NewTzKT(fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), "v1/events")),
		kinds:      tzktKinds,
		accounts:   accounts,
		api:        api.New(url),
		operations: make(chan OperationMessage, 1024),
		blocks:     make(chan BlockMessage, 1024),
	}
}

// Connect -
func (tzkt *TzKT) Connect(ctx context.Context) error {
	if err := tzkt.client.Connect(); err != nil {
		return err
	}

	tzkt.wg.Add(1)

	go func() {
		defer tzkt.wg.Done()

		for {
			select {
			case <-ctx.Done():
				if err := tzkt.close(); err != nil {
					log.Error(err)
				}
				return
			case msg := <-tzkt.client.Listen():
				switch msg.Channel {
				case events.ChannelOperations:
					if err := tzkt.handleOperationMessage(msg); err != nil {
						log.Error(err)
					}
				case events.ChannelBlocks:
					if err := tzkt.handleBlockMessage(msg); err != nil {
						log.Error(err)
					}
				}
			}
		}
	}()

	return nil
}

// Close -
func (tzkt *TzKT) Close() error {
	tzkt.wg.Wait()
	return nil
}

func (tzkt *TzKT) close() error {
	if err := tzkt.client.Close(); err != nil {
		return err
	}

	close(tzkt.operations)
	close(tzkt.blocks)
	return nil
}

// Operations -
func (tzkt *TzKT) Operations() <-chan OperationMessage {
	return tzkt.operations
}

// Operations -
func (tzkt *TzKT) Blocks() <-chan BlockMessage {
	return tzkt.blocks
}

func (tzkt *TzKT) handleBlockMessage(msg events.Message) error {
	switch msg.Type {
	case events.MessageTypeData:
		if msg.Body == nil {
			return nil
		}
		blocks := msg.Body.([]events.Block)
		for i := range blocks {
			tzkt.blocks <- BlockMessage{
				Hash:  blocks[i].Hash,
				Level: blocks[i].Level,
				Type:  msg.Type,
			}
		}
	case events.MessageTypeState, events.MessageTypeReorg, events.MessageTypeSubscribed:
		tzkt.blocks <- BlockMessage{
			Level: msg.State,
			Type:  msg.Type,
		}
	default:
		return errors.Wrapf(ErrUnknownMessageType, "%d", msg.Type)
	}

	return nil
}

func (tzkt *TzKT) handleOperationMessage(msg events.Message) error {
	switch msg.Type {
	case events.MessageTypeData:
		if msg.Body == nil {
			return nil
		}
		operations := msg.Body.([]interface{})
		return tzkt.handleUpdateMessage(operations)
	case events.MessageTypeState, events.MessageTypeReorg:
	default:
		return errors.Wrapf(ErrUnknownMessageType, "%d", msg.Type)
	}

	return nil
}

func (tzkt *TzKT) handleUpdateMessage(operations []interface{}) error {
	message := newOperationMessage()

	for i := range operations {
		if err := tzkt.processOperation(operations[i], &message); err != nil {
			return err
		}
	}

	tzkt.operations <- message
	tzkt.state = message.Level
	return nil
}

func (tzkt *TzKT) getAPIOperation(data interface{}) (api.Operation, error) {
	switch operation := data.(type) {

	case *events.Delegation:
		tx := api.Operation{
			ID:       operation.ID,
			Level:    operation.Level,
			Hash:     operation.Hash,
			Kind:     toNodeKinds[operation.Type],
			Block:    operation.Block,
			GasUsed:  &operation.GasUsed,
			BakerFee: &operation.BakerFee,
		}
		if operation.NewDelegate != nil {
			tx.Delegate = &api.Address{
				Alias:   operation.NewDelegate.Alias,
				Address: operation.NewDelegate.Address,
			}
		}
		return tx, nil

	case *events.Origination:
		tx := api.Operation{
			ID:       operation.ID,
			Level:    operation.Level,
			Hash:     operation.Hash,
			Kind:     toNodeKinds[operation.Type],
			Block:    operation.Block,
			GasUsed:  &operation.GasUsed,
			BakerFee: &operation.BakerFee,
		}
		return tx, nil

	case *events.Reveal:
		tx := api.Operation{
			ID:       operation.ID,
			Level:    operation.Level,
			Hash:     operation.Hash,
			Kind:     toNodeKinds[operation.Type],
			Block:    operation.Block,
			GasUsed:  &operation.GasUsed,
			BakerFee: &operation.BakerFee,
		}
		return tx, nil

	case *events.Transaction:
		tx := api.Operation{
			ID:       operation.ID,
			Level:    operation.Level,
			Hash:     operation.Hash,
			Kind:     toNodeKinds[operation.Type],
			Block:    operation.Block,
			GasUsed:  &operation.GasUsed,
			BakerFee: &operation.BakerFee,
		}
		if operation.Parameter != nil {
			tx.Parameters = &api.Parameters{
				Entrypoint: operation.Parameter.Entrypoint,
				Value:      operation.Parameter.Value,
			}
		}
		return tx, nil

	case map[string]interface{}:
		var general api.Operation
		err := mapstructure.Decode(data, &general)
		return general, err
	default:
		return api.Operation{}, errors.Wrapf(ErrInvalidOperationType, "%T", data)
	}
}

func (tzkt *TzKT) processOperation(data interface{}, message *OperationMessage) error {
	operation, err := tzkt.getAPIOperation(data)
	if err != nil {
		return err
	}
	operation.Kind = toNodeKinds[operation.Kind]
	if value, ok := message.Hash.LoadOrStore(operation.Hash, operation); ok {
		if stored, ok := value.(api.Operation); ok {
			if operation.BakerFee != nil {
				if stored.BakerFee != nil {
					*stored.BakerFee += *operation.BakerFee
				} else {
					stored.BakerFee = operation.BakerFee
				}
			}
			if operation.GasUsed != nil {
				if stored.GasUsed != nil {
					*stored.GasUsed += *operation.GasUsed
				} else {
					stored.GasUsed = operation.GasUsed
				}
			}
			message.Hash.Store(operation.Hash, stored)
		}
	}

	if message.Level == 0 {
		message.Level = operation.Level
		message.Block = operation.Block
	}
	return nil
}

// SubscribeToOperations - Sends operations of specified `types` or related to specified `account`, included into the blockchain
func (tzkt *TzKT) SubscribeToOperations(address string, types ...string) error {
	return tzkt.client.SubscribeToOperations(address, types...)
}

// SubscribeToBlocks -
func (tzkt *TzKT) SubscribeToBlocks() error {
	return tzkt.client.SubscribeToBlocks()
}

type tableState struct {
	Table    string
	LastID   uint64
	Finished bool
	Items    []api.Operation
}

func emptyTableState(table string) *tableState {
	return &tableState{
		Table: table,
		Items: make([]api.Operation, 0),
	}
}

type syncState []*tableState

func (a syncState) Len() int { return len(a) }
func (a syncState) Less(i, j int) bool {
	if !a[i].Finished && a[j].Finished {
		return true
	}

	switch {
	case len(a[i].Items) == 0 && len(a[j].Items) == 0:
		return false
	case len(a[i].Items) != 0 && len(a[j].Items) == 0:
		return false
	case len(a[i].Items) == 0 && len(a[j].Items) != 0:
		return true
	default:
		return a[i].Items[0].Level < a[j].Items[0].Level
	}
}
func (a syncState) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func newSyncState(kind ...string) syncState {
	ss := make(syncState, 0)
	if len(kind) == 0 {
		ss = append(ss, emptyTableState(api.KindTransaction))
	} else {
		for i := range kind {
			ss = append(ss, emptyTableState(kind[i]))
		}
	}
	return ss
}

func (state syncState) finished() bool {
	for i := range state {
		if !state[i].Finished {
			return false
		}
	}
	return true
}

func (state syncState) nextToRequest() *tableState {
	for i := range state {
		if !state[i].Finished && len(state[i].Items) == 0 {
			return state[i]
		}
	}
	return nil
}

// Sync -
func (tzkt *TzKT) Sync(ctx context.Context, indexerLevel uint64) {
	tzkt.state = indexerLevel

	head, err := tzkt.api.GetHead(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("current TzKT level is %d. Current mempool indexer level is %d", head.Level, tzkt.state)
	for {
		select {
		case <-ctx.Done():
			if err := tzkt.close(); err != nil {
				log.Error(err)
			}
			return
		default:
			if head.Level <= tzkt.state || tzkt.state == 0 {
				log.Info("synced")
				return
			}
			state := newSyncState(tzkt.kinds...)

			if len(state) == 0 {
				log.Error(ErrEmptyKindList)
				return
			}

			if err := tzkt.init(ctx, state, tzkt.state, head.Level); err != nil {
				log.Error(err)
				return
			}

			tzkt.state = head.Level
			head, err = tzkt.api.GetHead(ctx)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func (tzkt *TzKT) init(ctx context.Context, state syncState, indexerState, headLevel uint64) error {
	msg := newOperationMessage()

	for !state.finished() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			for table := state.nextToRequest(); table != nil; table = state.nextToRequest() {
				if err := tzkt.getTableData(ctx, table, indexerState, headLevel); err != nil {
					return err
				}
			}

			sort.Sort(state)

			if err := tzkt.processSync(ctx, state, &msg); err != nil {
				return err
			}
		}
	}

	return nil
}

// Subscribe -
func (tzkt *TzKT) Subscribe() error {
	if err := tzkt.SubscribeToBlocks(); err != nil {
		return err
	}

	if len(tzkt.accounts) == 0 {
		return tzkt.SubscribeToOperations("", tzkt.kinds...)
	}

	for _, account := range tzkt.accounts {
		if err := tzkt.SubscribeToOperations(account, tzkt.kinds...); err != nil {
			return err
		}
	}
	return nil
}

func (tzkt *TzKT) processSync(ctx context.Context, state syncState, msg *OperationMessage) error {
	for len(state[0].Items) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			table := state[0]

			operation := table.Items[0]
			msg.Hash.LoadOrStore(operation.Hash, operation)
			table.LastID = operation.ID

			switch {
			case msg.Level == 0:
				msg.Level = operation.Level
				msg.Block = operation.Block
			case msg.Level != operation.Level:
				tzkt.blocks <- BlockMessage{
					Type:  events.MessageTypeData,
					Level: msg.Level,
					Hash:  msg.Block,
				}
				tzkt.operations <- msg.copy()
				msg.clear()
			}

			table.Items = table.Items[1:]
			sort.Sort(state)
		}
	}

	if msg.Level > 0 && state.finished() {
		tzkt.blocks <- BlockMessage{
			Type:  events.MessageTypeData,
			Level: msg.Level,
			Hash:  msg.Block,
		}
		tzkt.operations <- msg.copy()
		msg.clear()
	}
	return nil
}

func (tzkt *TzKT) getTableData(ctx context.Context, table *tableState, indexerState, headLevel uint64) error {
	filters := map[string]string{
		"limit":         fmt.Sprintf("%d", pageSize),
		"level.le":      fmt.Sprintf("%d", headLevel),
		"select.fields": "hash,block,level,gasUsed,bakerFee,id",
	}

	if table.LastID == 0 {
		filters["level.gt"] = fmt.Sprintf("%d", indexerState)
	} else {
		filters["offset.cr"] = fmt.Sprintf("%d", table.LastID)
	}

	switch table.Table {
	case api.KindActivation:
		return getOperations(ctx, table, filters, tzkt.api.GetActivations)
	case api.KindBallot:
		return getOperations(ctx, table, filters, tzkt.api.GetBallots)
	case api.KindDelegation:
		return getOperations(ctx, table, filters, tzkt.api.GetDelegations)
	case api.KindDoubleBaking:
		return getOperations(ctx, table, filters, tzkt.api.GetDoubleBakings)
	case api.KindDoubleEndorsing:
		return getOperations(ctx, table, filters, tzkt.api.GetDoubleEndorsings)
	case api.KindEndorsement:
		return getOperations(ctx, table, filters, tzkt.api.GetEndorsements)
	case api.KindNonceRevelation:
		return getOperations(ctx, table, filters, tzkt.api.GetNonceRevelations)
	case api.KindOrigination:
		return getOperations(ctx, table, filters, tzkt.api.GetOriginations)
	case api.KindProposal:
		return getOperations(ctx, table, filters, tzkt.api.GetProposals)
	case api.KindReveal:
		return getOperations(ctx, table, filters, tzkt.api.GetReveals)
	case api.KindTransaction:
		return getOperations(ctx, table, filters, tzkt.api.GetTransactions)
	case api.KindRegisterGlobalConstant:
		return getOperations(ctx, table, filters, tzkt.api.GetRegisterConstants)
	default:
		return errors.Wrap(ErrUnknownOperationKind, table.Table)
	}
}

func getOperations(ctx context.Context, table *tableState, filters map[string]string, requestFunc func(context.Context, map[string]string) ([]api.Operation, error)) error {
	operations, err := requestFunc(ctx, filters)
	if err != nil {
		return err
	}
	if len(operations) != pageSize {
		table.Finished = true
	}
	for i := range operations {
		operations[i].Kind = table.Table
		table.Items = append(table.Items, operations[i])
	}
	return nil
}

// GetBlocks -
func (tzkt *TzKT) GetBlocks(ctx context.Context, limit, state uint64) ([]BlockMessage, error) {
	filters := map[string]string{
		"sort.desc":     "level",
		"limit":         fmt.Sprintf("%d", limit),
		"level.le":      fmt.Sprintf("%d", state),
		"select.fields": "hash,level",
	}
	blocks, err := tzkt.api.GetBlocks(ctx, filters)
	if err != nil {
		return nil, err
	}
	messages := make([]BlockMessage, 0, len(blocks))
	for i := range blocks {
		messages = append(messages, BlockMessage{
			Type:  events.MessageTypeData,
			Hash:  blocks[i].Hash,
			Level: blocks[i].Level,
		})
	}
	return messages, nil
}

// Delegates -
func (tzkt *TzKT) Delegates(ctx context.Context, limit, offset int64) ([]api.Delegate, error) {
	return tzkt.api.GetDelegates(ctx, map[string]string{
		"active": "true",
		"select": "publicKey,address",
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
}

// Rights -
func (tzkt *TzKT) Rights(ctx context.Context, level uint64) ([]api.Right, error) {
	return tzkt.api.GetRights(ctx, map[string]string{
		"type":   "endorsing",
		"level":  strconv.FormatUint(level, 10),
		"select": "baker,status,slots",
	})
}
