package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/api"
	"github.com/dipdup-net/go-lib/tzkt/events"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
	"github.com/dipdup-net/mempool/cmd/mempool/tzkt"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (indexer *Indexer) handleBlock(ctx context.Context, block tzkt.BlockMessage) error {
	if err := indexer.processOldOperations(); err != nil {
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
				if err := indexer.branches.Add(blocks[i]); err != nil {
					return err
				}
			}
		}
		return nil
	case events.MessageTypeData:
		if block.Level > indexer.state.Level {
			indexer.state.Level = block.Level
			indexer.log().Info("indexer state was updated")
			if err := indexer.state.Update(indexer.db); err != nil {
				return err
			}
		}
	}
	return indexer.branches.Add(block)
}

func (indexer *Indexer) processOldOperations() error {
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		if err := models.DeleteOldOperations(tx, indexer.keepInChain, models.StatusInChain, indexer.filters.Kinds...); err != nil {
			return errors.Wrap(err, "DeleteOldOperations in_chain")
		}
		if err := models.DeleteOldOperations(tx, indexer.keepOperations, "", indexer.filters.Kinds...); err != nil {
			return errors.Wrap(err, "DeleteOldOperations")
		}
		if indexer.hasManager {
			if err := models.DeleteOldGasStats(tx, indexer.gasStatsLifetime); err != nil {
				return errors.Wrap(err, "DeleteOldGasStats")
			}
		}
		return nil
	})
}

func (indexer *Indexer) handleInChain(operations tzkt.OperationMessage) error {
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		operations.Hash.Range(func(_, operation interface{}) bool {
			apiOperation, ok := operation.(api.Operation)
			if !ok {
				return false
			}
			if err := models.SetInChain(tx, indexer.network, apiOperation.Hash, apiOperation.Kind, operations.Level); err != nil {
				indexer.log().Error(err)
				return false
			}

			if indexer.prom != nil {
				indexer.prom.IncrementCounter(operationCountMetricName, map[string]string{
					"kind":    apiOperation.Kind,
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
					indexer.log().Error(err)
					return false
				}
			}

			return true
		})
		return nil
	})
}

func (indexer *Indexer) handleFailedOperation(operation node.FailedMonitor, status, protocol string) error {
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		for i := range operation.Contents {
			mempoolOperation := models.MempoolOperation{
				Network:   indexer.network,
				Status:    status,
				Hash:      operation.Hash,
				Branch:    operation.Branch,
				Signature: operation.Signature,
				Errors:    datatypes.JSON(operation.Error),
				Raw:       datatypes.JSON(operation.Raw),
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
	})
}

func (indexer *Indexer) handleAppliedOperation(operation node.Applied, protocol string) error {
	return indexer.db.Transaction(func(tx *gorm.DB) error {
		for i := range operation.Contents {
			mempoolOperation := models.MempoolOperation{
				Network:   indexer.network,
				Status:    models.StatusApplied,
				Hash:      operation.Hash,
				Branch:    operation.Branch,
				Signature: operation.Signature,
				Raw:       datatypes.JSON(operation.Raw),
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
	})
}

func (indexer *Indexer) handleContent(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	operation.Kind = content.Kind
	if indexer.prom != nil {
		indexer.prom.IncrementCounter(operationCountMetricName, map[string]string{
			"kind":    content.Kind,
			"status":  operation.Status,
			"network": indexer.network,
		})
	}

	switch content.Kind {
	case node.KindActivation:
		return handleActivateAccount(db, content, operation, indexer.filters.Accounts...)
	case node.KindBallot:
		return handleBallot(db, content, operation)
	case node.KindDelegation:
		return handleDelegation(db, content, operation)
	case node.KindDoubleBaking:
		return handleDoubleBaking(db, content, operation)
	case node.KindDoubleEndorsing:
		return handleDoubleEndorsing(db, content, operation)
	case node.KindEndorsement:
		return handleEndorsement(db, content, operation)
	case node.KindEndorsementWithSlot:
		return handleEndorsementWithSlot(db, content, operation)
	case node.KindNonceRevelation:
		return handleNonceRevelation(db, content, operation)
	case node.KindOrigination:
		return handleOrigination(db, content, operation)
	case node.KindProposal:
		return handleProposal(db, content, operation)
	case node.KindReveal:
		return handleReveal(db, content, operation, indexer.filters.Accounts...)
	case node.KindTransaction:
		return handleTransaction(db, content, operation, indexer.filters.Accounts...)
	case node.KindRegisterGlobalConstant:
		return handleRegisterGloabalConstant(db, content, operation)
	default:
		return errors.Wrap(node.ErrUnknownKind, content.Kind)
	}
}

func handleEndorsement(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var endorsement models.Endorsement
	if err := json.Unmarshal(content.Body, &endorsement); err != nil {
		return err
	}
	endorsement.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&endorsement).Error
}

func handleEndorsementWithSlot(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
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
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&endorsement).Error
}

func handleActivateAccount(db *gorm.DB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var activateAccount models.ActivateAccount
	if err := json.Unmarshal(content.Body, &activateAccount); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == activateAccount.Pkh {
				activateAccount.MempoolOperation = operation
				return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&activateAccount).Error
			}
		}
		return nil
	}

	activateAccount.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&activateAccount).Error
}

func handleTransaction(db *gorm.DB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var transaction models.Transaction
	if err := json.Unmarshal(content.Body, &transaction); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == transaction.Source || account == transaction.Destination {
				transaction.MempoolOperation = operation
				return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&transaction).Error
			}
		}
		return nil
	}

	transaction.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&transaction).Error
}

func handleReveal(db *gorm.DB, content node.Content, operation models.MempoolOperation, accounts ...string) error {
	var reveal models.Reveal
	if err := json.Unmarshal(content.Body, &reveal); err != nil {
		return err
	}

	if len(accounts) > 0 {
		for _, account := range accounts {
			if account == reveal.Source {
				reveal.MempoolOperation = operation
				return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&reveal).Error
			}
		}
		return nil
	}

	reveal.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&reveal).Error
}

func handleBallot(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var ballot models.Ballot
	if err := json.Unmarshal(content.Body, &ballot); err != nil {
		return err
	}
	ballot.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&ballot).Error
}

func handleDelegation(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var delegation models.Delegation
	if err := json.Unmarshal(content.Body, &delegation); err != nil {
		return err
	}
	delegation.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&delegation).Error
}

func handleDoubleBaking(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var doubleBaking models.DoubleBaking
	if err := json.Unmarshal(content.Body, &doubleBaking); err != nil {
		return err
	}
	doubleBaking.Fill()
	doubleBaking.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&doubleBaking).Error
}

func handleDoubleEndorsing(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var doubleEndorsing models.DoubleEndorsing
	if err := json.Unmarshal(content.Body, &doubleEndorsing); err != nil {
		return err
	}
	doubleEndorsing.Fill()
	doubleEndorsing.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&doubleEndorsing).Error
}

func handleNonceRevelation(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var nonceRevelation models.NonceRevelation
	if err := json.Unmarshal(content.Body, &nonceRevelation); err != nil {
		return err
	}
	nonceRevelation.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&nonceRevelation).Error
}

func handleOrigination(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var origination models.Origination
	if err := json.Unmarshal(content.Body, &origination); err != nil {
		return err
	}
	origination.Fill()
	origination.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&origination).Error
}

type proposals struct {
	Period    int64    `json:"period"`
	Proposals []string `json:"proposals"`
}

func handleProposal(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var proposal proposals
	if err := json.Unmarshal(content.Body, &proposal); err != nil {
		return err
	}
	for i := range proposal.Proposals {
		var p models.Proposal
		p.MempoolOperation = operation
		p.Proposals = proposal.Proposals[i]
		p.Period = proposal.Period
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func handleRegisterGloabalConstant(db *gorm.DB, content node.Content, operation models.MempoolOperation) error {
	var registerGlobalConstant models.RegisterGlobalConstant
	if err := json.Unmarshal(content.Body, &registerGlobalConstant); err != nil {
		return err
	}
	registerGlobalConstant.MempoolOperation = operation
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&registerGlobalConstant).Error
}

func (indexer *Indexer) isKindAvailiable(kind string) bool {
	for _, availiable := range indexer.filters.Kinds {
		if strings.HasPrefix(kind, availiable) {
			return true
		}
	}
	return false
}
