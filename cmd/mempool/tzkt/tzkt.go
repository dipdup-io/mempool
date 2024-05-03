package tzkt

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dipdup-io/workerpool"
	"github.com/dipdup-net/go-lib/tzkt/api"
	"github.com/dipdup-net/go-lib/tzkt/data"
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	g          workerpool.Group
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
		client:     events.NewTzKT(fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), "v1/ws")),
		kinds:      tzktKinds,
		accounts:   accounts,
		api:        api.New(url),
		operations: make(chan OperationMessage, 1024),
		blocks:     make(chan BlockMessage, 1024),
		g:          workerpool.NewGroup(),
	}
}

// Connect -
func (tzkt *TzKT) Connect(ctx context.Context) error {
	if err := tzkt.client.Connect(ctx); err != nil {
		return err
	}

	tzkt.g.GoCtx(ctx, func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-tzkt.client.Listen():
				switch msg.Type {
				case events.MessageTypeData:
					switch msg.Channel {
					case events.ChannelOperations:
						if err := tzkt.handleOperationMessage(msg); err != nil {
							log.Err(err).Msg("handleOperationMessage")
						}
					case events.ChannelBlocks:
						if err := tzkt.handleBlockMessage(msg); err != nil {
							log.Err(err).Msg("handleBlockMessage")
						}
					}
				case events.MessageTypeState:
					if msg.Channel != events.ChannelBlocks {
						continue
					}

					if tzkt.state < msg.State {
						// if blocks was missed in some reason we should index missed blocks
						log.Warn().Uint64("old_state", tzkt.state).Uint64("new_level", msg.State).Msg("detect missed blocks. resync...")

						tzkt.Sync(ctx, msg.State)
					}
					tzkt.state = msg.State
				case events.MessageTypeReorg, events.MessageTypeSubscribed:
				}

			}
		}
	})

	return nil
}

// Close -
func (tzkt *TzKT) Close() error {
	tzkt.g.Wait()

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
	if msg.Body == nil {
		return nil
	}
	blocks := msg.Body.([]data.Block)
	for i := range blocks {
		tzkt.blocks <- BlockMessage{
			Hash:      blocks[i].Hash,
			Level:     blocks[i].Level,
			Type:      msg.Type,
			Timestamp: blocks[i].Timestamp.UTC(),
		}
		tzkt.state = blocks[i].Level
	}

	return nil
}

func (tzkt *TzKT) handleOperationMessage(msg events.Message) error {
	if msg.Body == nil {
		return nil
	}
	operations, ok := msg.Body.([]any)
	if !ok {
		return nil
	}
	return tzkt.handleUpdateMessage(operations)
}

func (tzkt *TzKT) handleUpdateMessage(operations []any) error {
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

func (tzkt *TzKT) getAPIOperation(model interface{}) (data.Operation, error) {
	switch operation := model.(type) {

	case *data.Delegation:
		return operationFromDelegation(*operation), nil

	case *data.Origination:
		return operationFromOrigination(*operation), nil

	case *data.Reveal:
		return operationFromReveal(*operation), nil

	case *data.Transaction:
		return operationFromTransaction(*operation), nil

	case *data.Activation:
		return operationFromActivation(*operation), nil

	case *data.Ballot:
		return operationFromBallot(*operation), nil

	case *data.DoubleBaking:
		return operationFromDoubleBaking(*operation), nil

	case *data.DoubleEndorsing:
		return operationFromDoubleEndorsing(*operation), nil

	case *data.Endorsement:
		return operationFromEndorsement(*operation), nil

	case *data.NonceRevelation:
		return operationFromNonceRevelation(*operation), nil

	case *data.RegisterConstant:
		return operationFromRegisterConstant(*operation), nil

	case *data.Proposal:
		return operationFromProposal(*operation), nil

	case *data.SetDepositsLimit:
		return operationFromSetDepositsLimit(*operation), nil

	case *data.Preendorsement:
		return operationFromPreendorsement(*operation), nil

	case *data.TransferTicket:
		return operationFromTransferTicket(*operation), nil

	case *data.TxRollupCommit:
		return operationFromTxRollupCommit(*operation), nil

	case *data.TxRollupDispatchTicket:
		return operationFromTxRollupDispatchTicket(*operation), nil

	case *data.TxRollupFinalizeCommitment:
		return operationFromTxRollupFinalizeCommitment(*operation), nil

	case *data.TxRollupOrigination:
		return operationFromTxRollupOrigination(*operation), nil

	case *data.TxRollupRejection:
		return operationFromTxRollupRejection(*operation), nil

	case *data.TxRollupRemoveCommitment:
		return operationFromTxRollupRemoveCommitment(*operation), nil

	case *data.TxRollupReturnBond:
		return operationFromTxRollupReturnBond(*operation), nil

	case *data.TxRollupSubmitBatch:
		return operationFromTxRollupSubmitBatch(*operation), nil

	case *data.VdfRevelation:
		return operationFromVdfRevelation(*operation), nil

	case *data.IncreasePaidStorage:
		return operationFromIncreasePaidStorage(*operation), nil

	case *data.UpdateConsensusKey:
		return operationFromUpdateConsensusKey(*operation), nil

	case *data.DrainDelegate:
		return operationFromDrainDelegate(*operation), nil

	case *data.SmartRollupAddMessage:
		return operationFromSrAddMessage(*operation), nil

	case *data.SmartRollupCement:
		return operationFromSrCement(*operation), nil

	case *data.SmartRollupExecute:
		return operationFromSrExecute(*operation), nil

	case *data.SmartRollupOriginate:
		return operationFromSrOriginate(*operation), nil

	case *data.SmartRollupPublish:
		return operationFromSrPublish(*operation), nil

	case *data.SmartRollupRecoverBond:
		return operationFromSrRecoverBond(*operation), nil

	case *data.SmartRollupRefute:
		return operationFromSrRefute(*operation), nil

	case *data.DalPublishCommitment:
		return operationFromDalPublishCommitment(*operation), nil

	default:
		return data.Operation{}, errors.Wrapf(ErrInvalidOperationType, "%T", model)
	}
}

func (tzkt *TzKT) processOperation(model interface{}, message *OperationMessage) error {
	operation, err := tzkt.getAPIOperation(model)
	if err != nil {
		return err
	}
	if value, ok := message.Hash.LoadOrStore(operation.Hash, operation); ok {
		if stored, ok := value.(data.Operation); ok {
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
	Items    []data.Operation
}

func emptyTableState(table string) *tableState {
	return &tableState{
		Table: table,
		Items: make([]data.Operation, 0),
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
		ss = append(ss, emptyTableState(data.KindTransaction))
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
		log.Err(err).Msg("get tzkt head")
		return
	}

	log.Info().Msgf("current TzKT level is %d. Current mempool indexer level is %d", head.Level, tzkt.state)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if head.Level <= tzkt.state || tzkt.state == 0 {
				log.Info().Msg("synced")
				return
			}
			state := newSyncState(tzkt.kinds...)

			if len(state) == 0 {
				log.Err(ErrEmptyKindList).Msg("tzkt.Sync")
				return
			}

			if err := tzkt.init(ctx, state, tzkt.state, head.Level); err != nil {
				log.Err(err).Msg("tzkt.Sync")
				return
			}

			tzkt.state = head.Level
			head, err = tzkt.api.GetHead(ctx)
			if err != nil {
				log.Err(err).Msg("tzkt.Sync")
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
			return nil
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
					Type:      events.MessageTypeData,
					Level:     msg.Level,
					Hash:      msg.Block,
					Timestamp: msg.Timestamp.UTC(),
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
	case data.KindActivation:
		return getOperations(ctx, table, filters, tzkt.api.GetActivations, operationFromActivation)
	case data.KindBallot:
		return getOperations(ctx, table, filters, tzkt.api.GetBallots, operationFromBallot)
	case data.KindDelegation:
		return getOperations(ctx, table, filters, tzkt.api.GetDelegations, operationFromDelegation)
	case data.KindDoubleBaking:
		return getOperations(ctx, table, filters, tzkt.api.GetDoubleBakings, operationFromDoubleBaking)
	case data.KindDoubleEndorsing:
		return getOperations(ctx, table, filters, tzkt.api.GetDoubleEndorsings, operationFromDoubleEndorsing)
	case data.KindEndorsement:
		return getOperations(ctx, table, filters, tzkt.api.GetEndorsements, operationFromEndorsement)
	case data.KindNonceRevelation:
		return getOperations(ctx, table, filters, tzkt.api.GetNonceRevelations, operationFromNonceRevelation)
	case data.KindOrigination:
		return getOperations(ctx, table, filters, tzkt.api.GetOriginations, operationFromOrigination)
	case data.KindProposal:
		return getOperations(ctx, table, filters, tzkt.api.GetProposals, operationFromProposal)
	case data.KindReveal:
		return getOperations(ctx, table, filters, tzkt.api.GetReveals, operationFromReveal)
	case data.KindTransaction:
		return getOperations(ctx, table, filters, tzkt.api.GetTransactions, operationFromTransaction)
	case data.KindRegisterGlobalConstant:
		return getOperations(ctx, table, filters, tzkt.api.GetRegisterConstants, operationFromRegisterConstant)
	case data.KindSetDepositsLimit:
		return getOperations(ctx, table, filters, tzkt.api.GetSetDepositsLimit, operationFromSetDepositsLimit)
	case data.KindPreendorsement:
		return getOperations(ctx, table, filters, tzkt.api.GetPreendorsement, operationFromPreendorsement)
	case data.KindRollupDispatchTickets:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupDispatchTicket, operationFromTxRollupDispatchTicket)
	case data.KindRollupFinalizeCommitment:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupFinalizeCommitment, operationFromTxRollupFinalizeCommitment)
	case data.KindRollupReturnBond:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupReturnBond, operationFromTxRollupReturnBond)
	case data.KindRollupSubmitBatch:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupSubmitBatch, operationFromTxRollupSubmitBatch)
	case data.KindTransferTicket:
		return getOperations(ctx, table, filters, tzkt.api.GetTransferTicket, operationFromTransferTicket)
	case data.KindTxRollupCommit:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupCommit, operationFromTxRollupCommit)
	case data.KindTxRollupOrigination:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupOrigination, operationFromTxRollupOrigination)
	case data.KindTxRollupRejection:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupRejection, operationFromTxRollupRejection)
	case data.KindTxRollupRemoveCommitment:
		return getOperations(ctx, table, filters, tzkt.api.GetTxRollupRemoveCommitment, operationFromTxRollupRemoveCommitment)
	case data.KindVdfRevelation:
		return getOperations(ctx, table, filters, tzkt.api.GetVdfRevelations, operationFromVdfRevelation)
	case data.KindIncreasePaidStorage:
		return getOperations(ctx, table, filters, tzkt.api.GetIncreasePaidStorage, operationFromIncreasePaidStorage)
	case data.KindUpdateConsensusKey:
		return getOperations(ctx, table, filters, tzkt.api.GetUpdateConsensusKey, operationFromUpdateConsensusKey)
	case data.KindDrainDelegate:
		return getOperations(ctx, table, filters, tzkt.api.GetDrainDelegates, operationFromDrainDelegate)
	case data.KindSrAddMessages:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupAddMessages, operationFromSrAddMessage)
	case data.KindSrCement:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupCement, operationFromSrCement)
	case data.KindSrExecute:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupExecute, operationFromSrExecute)
	case data.KindSrOriginate:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupOriginate, operationFromSrOriginate)
	case data.KindSrPublish:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupPublish, operationFromSrPublish)
	case data.KindSrRecoverBond:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupRecoverBond, operationFromSrRecoverBond)
	case data.KindSrRefute:
		return getOperations(ctx, table, filters, tzkt.api.GetSmartRollupRefute, operationFromSrRefute)
	case data.KindDalPublishCommitment:
		return getOperations(ctx, table, filters, tzkt.api.GetDalPublishCommitment, operationFromDalPublishCommitment)
	default:
		return errors.Wrap(ErrUnknownOperationKind, table.Table)
	}
}

func getOperations[M data.OperationConstraint](ctx context.Context, table *tableState, filters map[string]string, receiver receiver[M], processor processor[M]) error {
	operations, err := receiver(ctx, filters)
	if err != nil {
		return err
	}
	if len(operations) != pageSize {
		table.Finished = true
	}
	for i := range operations {
		table.Items = append(table.Items, processor(operations[i]))
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
func (tzkt *TzKT) Delegates(ctx context.Context, limit, offset int64) ([]data.Delegate, error) {
	return tzkt.api.GetDelegates(ctx, map[string]string{
		"active": "true",
		"select": "publicKey,address",
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
}

// Rights -
func (tzkt *TzKT) Rights(ctx context.Context, level uint64) ([]data.Right, error) {
	return tzkt.api.GetRights(ctx, map[string]string{
		"type":   "endorsing",
		"level":  strconv.FormatUint(level, 10),
		"select": "baker,status,slots",
	})
}
