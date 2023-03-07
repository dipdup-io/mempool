package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/data"
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	"github.com/go-pg/pg/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
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
			if err := indexer.db.UpdateState(indexer.state); err != nil {
				return err
			}
		}
	}
	return indexer.branches.Add(ctx, block)
}

func (indexer *Indexer) handleOldOperations(ctx context.Context) error {
	return indexer.db.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		return indexer.processOldOperations(tx)
	})
}

func (indexer *Indexer) processOldOperations(db pg.DBI) error {
	if err := models.DeleteOldOperations(db, indexer.keepInChain, models.StatusInChain, indexer.filters.Kinds...); err != nil {
		return errors.Wrap(err, "DeleteOldOperations in_chain")
	}
	if err := models.DeleteOldOperations(db, indexer.keepOperations, "", indexer.filters.Kinds...); err != nil {
		return errors.Wrap(err, "DeleteOldOperations")
	}
	if indexer.hasManager {
		if err := models.DeleteOldGasStats(db, indexer.gasStatsLifetime); err != nil {
			return errors.Wrap(err, "DeleteOldGasStats")
		}
	}
	return nil
}

func (indexer *Indexer) handleInChain(ctx context.Context, operations tzkt.OperationMessage) error {
	return indexer.db.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		return indexer.inChainOperationProcess(tx, operations)
	})
}

func (indexer *Indexer) inChainOperationProcess(tx pg.DBI, operations tzkt.OperationMessage) error {
	operations.Hash.Range(func(_, operation interface{}) bool {
		apiOperation, ok := operation.(data.Operation)
		if !ok {
			return false
		}
		if err := models.SetInChain(tx, indexer.network, apiOperation.Hash, apiOperation.Type, operations.Level); err != nil {
			indexer.error().Err(err).Msg("models.SetInChain")
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
			if err := gasStats.Save(tx); err != nil {
				indexer.error().Err(err).Msg("gasStats.Save")
				return false
			}
		}

		return true
	})
	return nil
}

func (indexer *Indexer) handleFailedOperation(ctx context.Context, operation node.FailedMonitor, status, protocol string) error {
	return indexer.db.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		return indexer.failedOperationProcess(tx, operation, status, protocol)
	})
}

func (indexer *Indexer) failedOperationProcess(tx pg.DBI, operation node.FailedMonitor, status, protocol string) error {
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
		if err := indexer.handleContent(tx, operation.Contents[i], mempoolOperation); err != nil {
			return err
		}
	}
	return nil
}

func (indexer *Indexer) handleAppliedOperation(ctx context.Context, operation node.Applied, protocol string) error {
	return indexer.db.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		return indexer.appliedOperationProcess(tx, operation, protocol)
	})
}

func (indexer *Indexer) appliedOperationProcess(tx pg.DBI, operation node.Applied, protocol string) error {
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
		if err := indexer.handleContent(tx, operation.Contents[i], mempoolOperation); err != nil {
			return err
		}

		if indexer.hasManager {
			key := fmt.Sprintf("gas:%s", operation.Hash)
			if !indexer.cache.Has(key) {
				indexer.cache.Set(key)
				gasStats := models.GasStats{
					Network:        indexer.network,
					Hash:           operation.Hash,
					LevelInMempool: indexer.state.Level,
				}
				if err := gasStats.Save(tx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (indexer *Indexer) handleContent(tx pg.DBI, content node.Content, operation models.MempoolOperation) error {
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
		return handleActivateAccount(tx, content, operation, addresses...)
	case node.KindBallot:
		var model models.Ballot
		return defaultHandler(tx, content, operation, &model)
	case node.KindDelegation:
		var model models.Delegation
		return defaultHandler(tx, content, operation, &model)
	case node.KindDoubleBaking:
		return handleDoubleBaking(tx, content, operation)
	case node.KindDoubleEndorsing:
		return handleDoubleEndorsing(tx, content, operation)
	case node.KindEndorsement:
		return indexer.handleEndorsement(tx, content, operation)
	case node.KindEndorsementWithSlot:
		return indexer.handleEndorsementWithSlot(tx, content, operation)
	case node.KindNonceRevelation:
		var model models.NonceRevelation
		return defaultHandler(tx, content, operation, &model)
	case node.KindOrigination:
		return handleOrigination(tx, content, operation)
	case node.KindProposal:
		return handleProposal(tx, content, operation)
	case node.KindReveal:
		return handleReveal(tx, content, operation, addresses...)
	case node.KindTransaction:
		return handleTransaction(tx, content, operation, addresses...)
	case node.KindRegisterGlobalConstant:
		var model models.RegisterGlobalConstant
		return defaultHandler(tx, content, operation, &model)
	case node.KindDoublePreendorsement:
		var model models.DoublePreendorsing
		return defaultHandler(tx, content, operation, &model)
	case node.KindPreendorsement:
		var model models.Preendorsement
		return defaultHandler(tx, content, operation, &model)
	case node.KindSetDepositsLimit:
		return handleSetDepositsLimit(tx, content, operation, addresses...)
	case node.KindTransferTicket:
		var model models.TransferTicket
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupCommit:
		var model models.TxRollupCommit
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupDispatchTickets:
		var model models.TxRollupDispatchTickets
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupFinalizeCommitment:
		var model models.TxRollupFinalizeCommitment
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupOrigination:
		var model models.TxRollupOrigination
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupRejection:
		var model models.TxRollupRejection
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupRemoveCommitment:
		var model models.TxRollupRemoveCommitment
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupReturnBond:
		var model models.TxRollupReturnBond
		return defaultHandler(tx, content, operation, &model)
	case node.KindTxRollupSubmitBatch:
		var model models.TxRollupSubmitBatch
		return defaultHandler(tx, content, operation, &model)
	case node.KindIncreasePaidStorage:
		var model models.IncreasePaidStorage
		return defaultHandler(tx, content, operation, &model)
	case node.KindVdfRevelation:
		var model models.VdfRevelation
		return defaultHandler(tx, content, operation, &model)
	case node.KindUpdateConsensusKey:
		var model models.UpdateConsensusKey
		return defaultHandler(tx, content, operation, &model)
	case node.KindDrainDelegate:
		var model models.DelegateDrain
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrAddMessages:
		var model models.SmartRollupAddMessage
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrCement:
		var model models.SmartRollupCement
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrExecute:
		var model models.SmartRollupExecute
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrOriginate:
		var model models.SmartRollupOriginate
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrPublish:
		var model models.SmartRollupPublish
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrRecoverBond:
		var model models.SmartRollupRecoverBond
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrRefute:
		var model models.SmartRollupRefute
		return defaultHandler(tx, content, operation, &model)
	case node.KindSrTimeout:
		var model models.SmartRollupTimeout
		return defaultHandler(tx, content, operation, &model)
	case node.KindEvent:
	default:
		indexer.warn().Str("kind", content.Kind).Msg("unknown operation kind")
	}
	return nil
}

func createModel(db pg.DBI, model any) error {
	_, err := db.Model(model).OnConflict("DO NOTHING").Insert()
	return err
}

func (indexer *Indexer) handleEndorsement(db pg.DBI, content node.Content, operation models.MempoolOperation) error {
	var endorsement models.Endorsement
	if err := json.Unmarshal(content.Body, &endorsement); err != nil {
		return err
	}
	endorsement.MempoolOperation = operation

	if err := createModel(db, &endorsement); err != nil {
		return err
	}
	indexer.endorsements <- &endorsement
	return nil
}

func (indexer *Indexer) handleEndorsementWithSlot(db pg.DBI, content node.Content, operation models.MempoolOperation) error {
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

	if err := createModel(db, &endorsement); err != nil {
		return err
	}
	indexer.endorsements <- &endorsement
	return nil
}

func handleActivateAccount(db pg.DBI, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var activateAccount models.ActivateAccount
	if err := json.Unmarshal(content.Body, &activateAccount); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == activateAccount.Pkh {
				activateAccount.MempoolOperation = operation
				return createModel(db, &activateAccount)
			}
		}
		return nil
	}

	activateAccount.MempoolOperation = operation
	return createModel(db, &activateAccount)
}

func handleTransaction(tx pg.DBI, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var transaction models.Transaction
	if err := json.Unmarshal(content.Body, &transaction); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == transaction.Source || account == transaction.Destination {
				transaction.MempoolOperation = operation
				return createModel(tx, &transaction)
			}
		}
		return nil
	}

	transaction.MempoolOperation = operation
	return createModel(tx, &transaction)
}

func handleReveal(tx pg.DBI, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var reveal models.Reveal
	if err := json.Unmarshal(content.Body, &reveal); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == reveal.Source {
				reveal.MempoolOperation = operation
				return createModel(tx, &reveal)
			}
		}
		return nil
	}

	reveal.MempoolOperation = operation
	return createModel(tx, &reveal)
}

func handleDoubleBaking(tx pg.DBI, content node.Content, operation models.MempoolOperation) error {
	var doubleBaking models.DoubleBaking
	if err := json.Unmarshal(content.Body, &doubleBaking); err != nil {
		return err
	}
	doubleBaking.Fill()
	doubleBaking.MempoolOperation = operation
	return createModel(tx, &doubleBaking)
}

func handleDoubleEndorsing(tx pg.DBI, content node.Content, operation models.MempoolOperation) error {
	var doubleEndorsing models.DoubleEndorsing
	if err := json.Unmarshal(content.Body, &doubleEndorsing); err != nil {
		return err
	}
	doubleEndorsing.Fill()
	doubleEndorsing.MempoolOperation = operation
	return createModel(tx, &doubleEndorsing)
}

func handleOrigination(tx pg.DBI, content node.Content, operation models.MempoolOperation) error {
	var origination models.Origination
	if err := json.Unmarshal(content.Body, &origination); err != nil {
		return err
	}
	origination.Fill()
	origination.MempoolOperation = operation
	return createModel(tx, &origination)
}

type proposals struct {
	Period    int64    `json:"period"`
	Proposals []string `json:"proposals"`
}

func handleProposal(tx pg.DBI, content node.Content, operation models.MempoolOperation) error {
	var proposal proposals
	if err := json.Unmarshal(content.Body, &proposal); err != nil {
		return err
	}
	for i := range proposal.Proposals {
		var p models.Proposal
		p.MempoolOperation = operation
		p.Proposals = proposal.Proposals[i]
		p.Period = proposal.Period
		if err := createModel(tx, &p); err != nil {
			return err
		}
	}
	return nil
}

func handleSetDepositsLimit(tx pg.DBI, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var setDepositsLimit models.SetDepositsLimit
	if err := json.Unmarshal(content.Body, &setDepositsLimit); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == setDepositsLimit.Source {
				setDepositsLimit.MempoolOperation = operation
				return createModel(tx, &setDepositsLimit)
			}
		}
		return nil
	}

	setDepositsLimit.MempoolOperation = operation
	return createModel(tx, &setDepositsLimit)
}

func defaultHandler[M models.ChangableMempoolOperation](tx pg.DBI, content node.Content, operation models.MempoolOperation, model M) error {
	if err := json.Unmarshal(content.Body, model); err != nil {
		return err
	}
	model.SetMempoolOperation(operation)
	return createModel(tx, model)
}

func (indexer *Indexer) isKindAvailiable(kind string) bool {
	for _, availiable := range indexer.filters.Kinds {
		if strings.HasPrefix(kind, availiable) {
			return true
		}
	}
	return false
}
