package main

import (
	"context"
	"strings"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/data"
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (indexer *Indexer) handleBlock(ctx context.Context, block tzkt.BlockMessage) error {
	if err := indexer.handleOldOperations(ctx); err != nil {
		return err
	}

	switch block.Type {
	case events.MessageTypeState:
		if indexer.state.Level < block.Level {
			indexer.sync(ctx)
		}
		if space := indexer.branches.Space(); space > 0 {
			blocks, err := indexer.tzkt.GetBlocks(ctx, space, indexer.state.Level)
			if err != nil {
				return err
			}

			for i := len(blocks) - 1; i > -1; i-- {
				if err := indexer.branches.Add(ctx, blocks[i]); err != nil {
					return err
				}
			}
		}
		return nil
	case events.MessageTypeData:
		if block.Level > indexer.state.Level {
			indexer.state.Level = block.Level
			indexer.state.Hash = block.Hash
			indexer.state.Timestamp = block.Timestamp
			indexer.info().Msg("indexer state was updated")
			if err := indexer.db.UpdateState(ctx, indexer.state); err != nil {
				return err
			}
		}
	}
	return indexer.branches.Add(ctx, block)
}

func (indexer *Indexer) handleOldOperations(ctx context.Context) error {
	return indexer.db.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return indexer.processOldOperations(ctx, tx)
	})
}

func (indexer *Indexer) processOldOperations(ctx context.Context, db bun.IDB) error {
	if err := models.DeleteOldOperations(ctx, db, indexer.keepInChain, models.StatusInChain, indexer.filters.Kinds...); err != nil {
		return errors.Wrap(err, "DeleteOldOperations in_chain")
	}
	if err := models.DeleteOldOperations(ctx, db, indexer.keepOperations, "", indexer.filters.Kinds...); err != nil {
		return errors.Wrap(err, "DeleteOldOperations")
	}
	if indexer.hasManager {
		if err := models.DeleteOldGasStats(ctx, db, indexer.gasStatsLifetime); err != nil {
			return errors.Wrap(err, "DeleteOldGasStats")
		}
	}
	return nil
}

func (indexer *Indexer) handleInChain(ctx context.Context, operations tzkt.OperationMessage) error {
	return indexer.db.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return indexer.inChainOperationProcess(ctx, tx, operations)
	})
}

func (indexer *Indexer) inChainOperationProcess(ctx context.Context, tx bun.IDB, operations tzkt.OperationMessage) error {
	operations.Hash.Range(func(_, operation interface{}) bool {
		apiOperation, ok := operation.(data.Operation)
		if !ok {
			return false
		}
		if err := models.SetInChain(ctx, tx, indexer.network, apiOperation.Hash, apiOperation.Type, operations.Level); err != nil {
			indexer.error(err).Msg("models.SetInChain")
			return false
		}

		if indexer.prom != nil {
			indexer.prom.IncrementCounter(operationCountMetricName, map[string]string{
				"kind":    apiOperation.Type,
				"status":  models.StatusInChain,
				"network": indexer.network,
			})
		}

		if indexer.hasManager {
			gasStats := models.GasStats{
				Network:      indexer.network,
				Hash:         apiOperation.Hash,
				LevelInChain: operations.Level,
			}
			if apiOperation.GasUsed != nil {
				gasStats.TotalGasUsed = *apiOperation.GasUsed
			}
			if apiOperation.BakerFee != nil {
				gasStats.TotalFee = *apiOperation.BakerFee
			}
			if err := gasStats.Save(ctx, tx); err != nil {
				indexer.error(err).Msg("gasStats.Save")
				return false
			}
		}

		return true
	})
	return nil
}

func (indexer *Indexer) handleFailedOperation(ctx context.Context, operation node.FailedMonitor, status, protocol string) error {
	return indexer.db.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return indexer.failedOperationProcess(ctx, tx, operation, status, protocol)
	})
}

func (indexer *Indexer) failedOperationProcess(ctx context.Context, tx bun.IDB, operation node.FailedMonitor, status, protocol string) error {
	for i := range operation.Contents {
		mempoolOperation := models.MempoolOperation{
			Network:   indexer.network,
			Status:    status,
			Hash:      operation.Hash,
			Branch:    operation.Branch,
			Signature: operation.Signature,
			Errors:    models.JSONB(operation.Error),
			Raw:       models.JSONB(operation.Raw),
			Protocol:  protocol,
		}
		if !indexer.isKindAvailiable(operation.Contents[i].Kind) {
			continue
		}
		if err := indexer.handleContent(ctx, tx, operation.Contents[i], mempoolOperation); err != nil {
			return err
		}
	}
	return nil
}

func (indexer *Indexer) handleAppliedOperation(ctx context.Context, operation node.Applied, protocol string) error {
	return indexer.db.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return indexer.appliedOperationProcess(ctx, tx, operation, protocol)
	})
}

func (indexer *Indexer) appliedOperationProcess(ctx context.Context, tx bun.IDB, operation node.Applied, protocol string) error {
	for i := range operation.Contents {
		mempoolOperation := models.MempoolOperation{
			Network:   indexer.network,
			Status:    models.StatusApplied,
			Hash:      operation.Hash,
			Branch:    operation.Branch,
			Signature: operation.Signature,
			Raw:       models.JSONB(operation.Raw),
			Protocol:  protocol,
		}
		expirationLevel := indexer.branches.ExpirationLevel(operation.Branch)
		if expirationLevel > 0 {
			mempoolOperation.ExpirationLevel = &expirationLevel
		}
		if !indexer.isKindAvailiable(operation.Contents[i].Kind) {
			continue
		}
		if err := indexer.handleContent(ctx, tx, operation.Contents[i], mempoolOperation); err != nil {
			return err
		}

		if indexer.hasManager {
			gasStats := models.GasStats{
				Network:        indexer.network,
				Hash:           operation.Hash,
				LevelInMempool: indexer.state.Level,
			}
			if err := gasStats.Save(ctx, tx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (indexer *Indexer) handleContent(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	operation.Kind = content.Kind
	if indexer.prom != nil {
		indexer.prom.IncrementCounter(operationCountMetricName, map[string]string{
			"kind":    content.Kind,
			"status":  operation.Status,
			"network": indexer.network,
		})
	}

	addresses := indexer.filters.Addresses()

	switch content.Kind {
	case node.KindActivation:
		return handleActivateAccount(ctx, tx, content, operation, addresses...)
	case node.KindBallot:
		var model models.Ballot
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindDelegation:
		var model models.Delegation
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindDoubleBaking:
		return handleDoubleBaking(ctx, tx, content, operation)
	case node.KindDoubleEndorsing:
		return handleDoubleEndorsing(ctx, tx, content, operation)
	case node.KindEndorsement:
		return indexer.handleEndorsement(ctx, tx, content, operation)
	case node.KindEndorsementWithSlot:
		return indexer.handleEndorsementWithSlot(ctx, tx, content, operation)
	case node.KindEndorsementWithDal:
		return indexer.handleEndorsement(ctx, tx, content, operation)
	case node.KindNonceRevelation:
		var model models.NonceRevelation
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindOrigination:
		return handleOrigination(ctx, tx, content, operation)
	case node.KindProposal:
		return handleProposal(ctx, tx, content, operation)
	case node.KindReveal:
		return handleReveal(ctx, tx, content, operation, addresses...)
	case node.KindTransaction:
		return handleTransaction(ctx, tx, content, operation, addresses...)
	case node.KindRegisterGlobalConstant:
		var model models.RegisterGlobalConstant
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindDoublePreendorsement:
		var model models.DoublePreendorsing
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindPreendorsement:
		var model models.Preendorsement
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSetDepositsLimit:
		return handleSetDepositsLimit(ctx, tx, content, operation, addresses...)
	case node.KindTransferTicket:
		var model models.TransferTicket
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupCommit:
		var model models.TxRollupCommit
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupDispatchTickets:
		var model models.TxRollupDispatchTickets
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupFinalizeCommitment:
		var model models.TxRollupFinalizeCommitment
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupOrigination:
		var model models.TxRollupOrigination
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupRejection:
		var model models.TxRollupRejection
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupRemoveCommitment:
		var model models.TxRollupRemoveCommitment
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupReturnBond:
		var model models.TxRollupReturnBond
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindTxRollupSubmitBatch:
		var model models.TxRollupSubmitBatch
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindIncreasePaidStorage:
		var model models.IncreasePaidStorage
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindVdfRevelation:
		var model models.VdfRevelation
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindUpdateConsensusKey:
		var model models.UpdateConsensusKey
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindDrainDelegate:
		var model models.DelegateDrain
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrAddMessages:
		var model models.SmartRollupAddMessage
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrCement:
		var model models.SmartRollupCement
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrExecute:
		var model models.SmartRollupExecute
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrOriginate:
		var model models.SmartRollupOriginate
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrPublish:
		var model models.SmartRollupPublish
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrRecoverBond:
		var model models.SmartRollupRecoverBond
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrRefute:
		var model models.SmartRollupRefute
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindSrTimeout:
		var model models.SmartRollupTimeout
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindDalPublishCommitment:
		var model models.DalPublishCommitment
		return defaultHandler(ctx, tx, content, operation, &model)
	case node.KindEvent:
	default:
		indexer.warn().Str("kind", content.Kind).Msg("unknown operation kind")
	}
	return nil
}

func createModel(ctx context.Context, tx bun.IDB, model any) error {
	_, err := tx.NewInsert().Model(model).On("CONFLICT DO NOTHING").Exec(ctx)
	return err
}

func (indexer *Indexer) handleEndorsement(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var endorsement models.Endorsement
	if err := json.Unmarshal(content.Body, &endorsement); err != nil {
		return err
	}
	endorsement.MempoolOperation = operation

	if err := createModel(ctx, tx, &endorsement); err != nil {
		return err
	}
	indexer.endorsements <- &endorsement
	return nil
}

func (indexer *Indexer) handleEndorsementWithSlot(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var endorsementWithSlot node.EndorsementWithSlot
	if err := json.Unmarshal(content.Body, &endorsementWithSlot); err != nil {
		return err
	}
	operation.Kind = node.KindEndorsement
	operation.Signature = endorsementWithSlot.Endorsement.Signature
	endorsement := models.Endorsement{
		MempoolOperation: operation,
		Level:            endorsementWithSlot.Endorsement.Operation.Level,
	}

	if err := createModel(ctx, tx, &endorsement); err != nil {
		return err
	}
	indexer.endorsements <- &endorsement
	return nil
}

func handleActivateAccount(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var activateAccount models.ActivateAccount
	if err := json.Unmarshal(content.Body, &activateAccount); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == activateAccount.Pkh {
				activateAccount.MempoolOperation = operation
				return createModel(ctx, tx, &activateAccount)
			}
		}
		return nil
	}

	activateAccount.MempoolOperation = operation
	return createModel(ctx, tx, &activateAccount)
}

func handleTransaction(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var transaction models.Transaction
	if err := json.Unmarshal(content.Body, &transaction); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == transaction.Source || account == transaction.Destination {
				transaction.MempoolOperation = operation
				return createModel(ctx, tx, &transaction)
			}
		}
		return nil
	}

	transaction.MempoolOperation = operation
	return createModel(ctx, tx, &transaction)
}

func handleReveal(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var reveal models.Reveal
	if err := json.Unmarshal(content.Body, &reveal); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == reveal.Source {
				reveal.MempoolOperation = operation
				return createModel(ctx, tx, &reveal)
			}
		}
		return nil
	}

	reveal.MempoolOperation = operation
	return createModel(ctx, tx, &reveal)
}

func handleDoubleBaking(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var doubleBaking models.DoubleBaking
	if err := json.Unmarshal(content.Body, &doubleBaking); err != nil {
		return err
	}
	doubleBaking.Fill()
	doubleBaking.MempoolOperation = operation
	return createModel(ctx, tx, &doubleBaking)
}

func handleDoubleEndorsing(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var doubleEndorsing models.DoubleEndorsing
	if err := json.Unmarshal(content.Body, &doubleEndorsing); err != nil {
		return err
	}
	doubleEndorsing.Fill()
	doubleEndorsing.MempoolOperation = operation
	return createModel(ctx, tx, &doubleEndorsing)
}

func handleOrigination(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var origination models.Origination
	if err := json.Unmarshal(content.Body, &origination); err != nil {
		return err
	}
	origination.Fill()
	origination.MempoolOperation = operation
	return createModel(ctx, tx, &origination)
}

type proposals struct {
	Period    int64    `json:"period"`
	Proposals []string `json:"proposals"`
}

func handleProposal(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation) error {
	var proposal proposals
	if err := json.Unmarshal(content.Body, &proposal); err != nil {
		return err
	}
	for i := range proposal.Proposals {
		var p models.Proposal
		p.MempoolOperation = operation
		p.Proposals = proposal.Proposals[i]
		p.Period = proposal.Period
		if err := createModel(ctx, tx, &p); err != nil {
			return err
		}
	}
	return nil
}

func handleSetDepositsLimit(ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var setDepositsLimit models.SetDepositsLimit
	if err := json.Unmarshal(content.Body, &setDepositsLimit); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == setDepositsLimit.Source {
				setDepositsLimit.MempoolOperation = operation
				return createModel(ctx, tx, &setDepositsLimit)
			}
		}
		return nil
	}

	setDepositsLimit.MempoolOperation = operation
	return createModel(ctx, tx, &setDepositsLimit)
}

func defaultHandler[M models.ChangableMempoolOperation](ctx context.Context, tx bun.IDB, content node.Content, operation models.MempoolOperation, model M) error {
	if err := json.Unmarshal(content.Body, model); err != nil {
		return err
	}
	model.SetMempoolOperation(operation)
	return createModel(ctx, tx, model)
}

func (indexer *Indexer) isKindAvailiable(kind string) bool {
	for _, availiable := range indexer.filters.Kinds {
		if strings.HasPrefix(kind, availiable) {
			return true
		}
	}
	return false
}
