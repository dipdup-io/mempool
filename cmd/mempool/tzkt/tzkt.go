package tzkt

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	events "github.com/dipdup-net/tzktevents"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	pageSize = 1000
)

// TzKT - tzkt data source
type TzKT struct {
	api    *TzKTAPI
	client *events.TzKT
	state  uint64
	kinds  []string

	operations chan OperationMessage
	blocks     chan BlockMessage
	stop       chan struct{}
	wg         sync.WaitGroup
}

// NewTzKT - TzKT data source`s constructor
func NewTzKT(url string, kinds []string) *TzKT {
	return &TzKT{
		client:     events.NewTzKT(fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), "v1/events")),
		kinds:      kinds,
		api:        NewTzKTAPI(url),
		operations: make(chan OperationMessage, 1024),
		blocks:     make(chan BlockMessage, 1024),
		stop:       make(chan struct{}, 1),
	}
}

// Connect -
func (tzkt *TzKT) Connect() error {
	if err := tzkt.client.Connect(); err != nil {
		return err
	}

	tzkt.wg.Add(1)

	go func() {
		defer tzkt.wg.Done()

		for {
			select {
			case <-tzkt.stop:
				return
			case msg := <-tzkt.client.Listen():
				switch msg.Channel {
				case "operations":
					if err := tzkt.handleOperationMessage(msg); err != nil {
						log.Error(err)
					}
				case "blocks":
					if err := tzkt.handleBlockMessage(msg); err != nil {
						log.Error(err)
					}
				default:
					log.Errorf("Unknown channel %s", msg.Channel)
				}
			}
		}
	}()

	return nil
}

// Close -
func (tzkt *TzKT) Close() error {
	tzkt.stop <- struct{}{}
	tzkt.wg.Wait()

	if err := tzkt.client.Close(); err != nil {
		return err
	}

	close(tzkt.operations)
	close(tzkt.blocks)
	close(tzkt.stop)

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
	case events.MessageTypeState, events.MessageTypeReorg:
		tzkt.blocks <- BlockMessage{
			Level: msg.State,
			Type:  msg.Type,
		}
	case events.MessageTypeData:
		data, ok := msg.Body.([]interface{})
		if !ok {
			return errors.Wrapf(ErrInvalidBlockType, "%v", msg.Body)
		}
		if len(data) == 0 {
			return nil
		}
		blockData, ok := data[0].(map[string]interface{})
		if !ok {
			return errors.Wrapf(ErrInvalidBlockType, "%v", data[0])
		}
		hash, err := getString(blockData, "hash")
		if err != nil {
			return err
		}
		level, err := getUint64(blockData, "level")
		if err != nil {
			return err
		}
		tzkt.blocks <- BlockMessage{
			Hash:  hash,
			Level: level,
			Type:  msg.Type,
		}
	default:
		return errors.Wrapf(ErrUnknownMessageType, "%d", msg.Type)
	}

	return nil
}

func (tzkt *TzKT) handleOperationMessage(msg events.Message) error {
	switch msg.Type {
	case events.MessageTypeState:
	case events.MessageTypeData:
		return tzkt.handleUpdateMessage(msg)
	case events.MessageTypeReorg:
	default:
		return errors.Wrapf(ErrUnknownMessageType, "%d", msg.Type)
	}

	return nil
}

func (tzkt *TzKT) handleUpdateMessage(msg events.Message) error {
	message := newOperationMessage()

	body, ok := msg.Body.([]interface{})
	if !ok {
		return errors.Wrapf(ErrInvalidBodyType, "%T", msg.Body)
	}
	for i := range body {
		operation, ok := body[i].(map[string]interface{})
		if !ok {
			return errors.Wrapf(ErrInvalidOperationType, "%T", body[i])
		}
		if err := tzkt.processOperation(operation, &message); err != nil {
			return err
		}
	}

	tzkt.operations <- message
	tzkt.state = message.Level
	return nil
}

func getString(data map[string]interface{}, key string) (string, error) {
	value, ok := data[key]
	if !ok {
		return "", errors.Wrapf(ErrOperationDoesNotContain, "field=%s data=%v", key, data)
	}
	s, ok := value.(string)
	if !ok {
		return "", errors.Wrapf(ErrInvalidFieldType, "field=%s expected_type=string type=%T data=%v", key, value, value)
	}
	return s, nil
}

func getUint64(data map[string]interface{}, key string) (uint64, error) {
	value, ok := data[key]
	if !ok {
		return 0, errors.Wrapf(ErrOperationDoesNotContain, "field=%s data=%v", key, data)
	}
	f, ok := value.(float64)
	if !ok {
		return 0, errors.Wrapf(ErrInvalidFieldType, "field=%s expected_type=float64 type=%T data=%v", key, value, value)
	}
	return uint64(f), nil
}

func (tzkt *TzKT) processOperation(data map[string]interface{}, message *OperationMessage) error {
	hash, err := getString(data, "hash")
	if err != nil {
		return err
	}
	kind, err := getString(data, "type")
	if err != nil {
		return err
	}
	msgKind, ok := toNodeKinds[kind]
	if !ok {
		return errors.Wrap(ErrUnknownOperationKind, kind)
	}
	message.Hash.Store(hash, msgKind)

	if message.Level == 0 {
		level, err := getUint64(data, "level")
		if err != nil {
			return err
		}
		message.Level = level

		block, err := getString(data, "block")
		if err != nil {
			return err
		}
		message.Block = block
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
	Items    []Operation
}

func emptyTableState(table string) *tableState {
	return &tableState{
		Table: table,
		Items: make([]Operation, 0),
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
		ss = append(ss, emptyTableState(KindTransaction))
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
func (tzkt *TzKT) Sync(indexerLevel uint64) error {
	tzkt.state = indexerLevel

	head, err := tzkt.api.GetHead()
	if err != nil {
		return err
	}

	log.Infof("Current TzKT level is %d. Current mempool indexer level is %d", head.Level, tzkt.state)
	for head.Level > tzkt.state {
		state := newSyncState(tzkt.kinds...)

		if len(state) == 0 {
			return ErrEmptyKindList
		}

		if err := tzkt.init(state, tzkt.state, head.Level); err != nil {
			return err
		}

		tzkt.state = head.Level
		head, err = tzkt.api.GetHead()
		if err != nil {
			return err
		}
		log.Infof("Indexed to level %d", tzkt.state)
	}

	log.Info("Synced")

	return nil
}

func (tzkt *TzKT) init(state syncState, indexerState, headLevel uint64) error {
	msg := newOperationMessage()

	for !state.finished() {

		for table := state.nextToRequest(); table != nil; table = state.nextToRequest() {
			if err := tzkt.getTableData(table, indexerState, headLevel); err != nil {
				return err
			}
		}

		if err := tzkt.processSync(state, &msg); err != nil {
			return err
		}
	}

	return nil
}

func (tzkt *TzKT) processSync(state syncState, msg *OperationMessage) error {
	for len(state[0].Items) > 0 {
		sort.Sort(state)

		table := state[0]

		operation := table.Items[0]
		kind, ok := toNodeKinds[operation.Kind]
		if !ok {
			return errors.Wrap(ErrUnknownOperationKind, kind)
		}
		msg.Hash.LoadOrStore(operation.Hash, kind)
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
	}

	if msg.Level > 0 {
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

func (tzkt *TzKT) getTableData(table *tableState, indexerState, headLevel uint64) error {
	filters := map[string]string{
		"limit":         fmt.Sprintf("%d", pageSize),
		"level.le":      fmt.Sprintf("%d", headLevel),
		"select.fields": "hash,block,level,id",
	}

	if table.LastID == 0 {
		filters["level.gt"] = fmt.Sprintf("%d", indexerState)
	} else {
		filters["offset.cr"] = fmt.Sprintf("%d", table.LastID)
	}

	switch table.Table {
	case KindActivation:
		return getOperations(table, filters, tzkt.api.GetActivations)
	case KindBallot:
		return getOperations(table, filters, tzkt.api.GetBallots)
	case KindDelegation:
		return getOperations(table, filters, tzkt.api.GetDelegations)
	case KindDoubleBaking:
		return getOperations(table, filters, tzkt.api.GetDoubleBakings)
	case KindDoubleEndorsing:
		return getOperations(table, filters, tzkt.api.GetDoubleEndorsings)
	case KindEndorsement:
		return getOperations(table, filters, tzkt.api.GetEndorsements)
	case KindNonceRevelation:
		return getOperations(table, filters, tzkt.api.GetNonceRevelations)
	case KindOrigination:
		return getOperations(table, filters, tzkt.api.GetOriginations)
	case KindProposal:
		return getOperations(table, filters, tzkt.api.GetProposals)
	case KindReveal:
		return getOperations(table, filters, tzkt.api.GetReveals)
	case KindTransaction:
		return getOperations(table, filters, tzkt.api.GetTransactions)
	default:
		return errors.Wrap(ErrUnknownOperationKind, table.Table)
	}
}

func getOperations(table *tableState, filters map[string]string, requestFunc func(map[string]string) ([]Operation, error)) error {
	operations, err := requestFunc(filters)
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
func (tzkt *TzKT) GetBlocks(limit, state uint64) ([]BlockMessage, error) {
	filters := map[string]string{
		"sort.desc":     "level",
		"limit":         fmt.Sprintf("%d", limit),
		"level.le":      fmt.Sprintf("%d", state),
		"select.fields": "hash,level",
	}
	blocks, err := tzkt.api.GetBlocks(filters)
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
