package tzkt

import (
	"context"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tzkt/data"
)

type processor[O data.OperationConstraint] func(model O) data.Operation
type receiver[O data.OperationConstraint] func(ctx context.Context, filters map[string]string) ([]O, error)

func operationFromTransaction(model data.Transaction) data.Operation {
	return data.Operation{
		Type:       node.KindTransaction,
		Level:      model.Level,
		ID:         model.ID,
		Hash:       model.Hash,
		Block:      model.Block,
		GasUsed:    &model.GasUsed,
		BakerFee:   &model.BakerFee,
		Parameters: model.Parameter,
	}
}

func operationFromOrigination(model data.Origination) data.Operation {
	return data.Operation{
		Type:     node.KindOrigination,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
	}
}

func operationFromDelegation(model data.Delegation) data.Operation {
	return data.Operation{
		Type:     node.KindDelegation,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
		Delegate: model.NewDelegate,
	}
}

func operationFromReveal(model data.Reveal) data.Operation {
	return data.Operation{
		Type:     node.KindReveal,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
	}
}

func operationFromRegisterConstant(model data.RegisterConstant) data.Operation {
	return data.Operation{
		Type:     node.KindRegisterGlobalConstant,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
	}
}

func operationFromBallot(model data.Ballot) data.Operation {
	return data.Operation{
		Type:  node.KindBallot,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromProposal(model data.Proposal) data.Operation {
	return data.Operation{
		Type:  node.KindProposal,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromNonceRevelation(model data.NonceRevelation) data.Operation {
	return data.Operation{
		Type:  node.KindNonceRevelation,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromEndorsement(model data.Endorsement) data.Operation {
	return data.Operation{
		Type:  node.KindEndorsement,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromDoubleBaking(model data.DoubleBaking) data.Operation {
	return data.Operation{
		Type:  node.KindDoubleBaking,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromDoubleEndorsing(model data.DoubleEndorsing) data.Operation {
	return data.Operation{
		Type:  node.KindDoubleEndorsing,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromActivation(model data.Activation) data.Operation {
	return data.Operation{
		Type:  node.KindActivation,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromSetDepositsLimit(model data.SetDepositsLimit) data.Operation {
	return data.Operation{
		Type:     node.KindSetDepositsLimit,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
	}
}

func operationFromPreendorsement(model data.Preendorsement) data.Operation {
	return data.Operation{
		Type:  node.KindSetDepositsLimit,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupCommit(model data.TxRollupCommit) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupCommit,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupDispatchTicket(model data.TxRollupDispatchTicket) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupDispatchTickets,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupFinalizeCommitment(model data.TxRollupFinalizeCommitment) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupFinalizeCommitment,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupOrigination(model data.TxRollupOrigination) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupOrigination,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupRejection(model data.TxRollupRejection) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupRejection,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupRemoveCommitment(model data.TxRollupRemoveCommitment) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupRemoveCommitment,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupReturnBond(model data.TxRollupReturnBond) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupReturnBond,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTxRollupSubmitBatch(model data.TxRollupSubmitBatch) data.Operation {
	return data.Operation{
		Type:  node.KindTxRollupSubmitBatch,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromTransferTicket(model data.TransferTicket) data.Operation {
	return data.Operation{
		Type:  node.KindTransferTicket,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}

func operationFromIncreasePaidStorage(model data.IncreasePaidStorage) data.Operation {
	return data.Operation{
		Type:     node.KindIncreasePaidStorage,
		Level:    model.Level,
		ID:       model.ID,
		Hash:     model.Hash,
		Block:    model.Block,
		GasUsed:  &model.GasUsed,
		BakerFee: &model.BakerFee,
	}
}

func operationFromVdfRevelation(model data.VdfRevelation) data.Operation {
	return data.Operation{
		Type:  node.KindVdfRevelation,
		Level: model.Level,
		ID:    model.ID,
		Hash:  model.Hash,
		Block: model.Block,
	}
}
