package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Statuses
const (
	StatusApplied       = "applied"
	StatusBranchDelayed = "branch_delayed"
	StatusBranchRefused = "branch_refused"
	StatusRefused       = "refused"
	StatusInChain       = "in_chain"
	StatusExpired       = "expired"
)

// DefaultConstraint -
type DefaultConstraint interface {
	Ballot | ActivateAccount | Delegation | DoubleBaking | DoubleEndorsing | DoublePreendorsing | Endorsement | 
		NonceRevelation | Origination | Preendorsement | Proposal | RegisterGlobalConstant | Reveal | SetDepositsLimit |
		Transaction | TransferTicket | TxRollupCommit | TxRollupDispatchTickets | TxRollupFinalizeCommitment | TxRollupOrigination |
		TxRollupRejection | TxRollupRemoveCommitment | TxRollupReturnBond | TxRollupSubmitBatch | VdfRevelation | IncreasePaidStorage |
		UpdateConsensusKey | DelegateDrain | SmartRollupAddMessage | SmartRollupCement | SmartRollupExecute |
		SmartRollupOriginate | SmartRollupPublish | SmartRollupRecoverBond | SmartRollupRefute | SmartRollupTimeout | DalPublishCommitment
}

// ChangableMempoolOperation -
type ChangableMempoolOperation interface {
	SetMempoolOperation(operation MempoolOperation)
}

// MempoolOperation -
type MempoolOperation struct {
	CreatedAt       int64   `comment:"Date of creation in seconds since UNIX epoch."                                                 json:"-"`
	UpdatedAt       int64   `comment:"Date of last update in seconds since UNIX epoch."                                              json:"-"`
	Network         string  `bun:",pk"                                                                                               comment:"Identifies belonging network."                json:"network"`
	Hash            string  `bun:",pk"                                                                                               comment:"Hash of the operation."                       json:"hash"`
	Branch          string  `comment:"Hash of the block, in which the operation was included."                                       json:"branch"`
	Status          string  `comment:"Status of the operation."                                                                      json:"status"`
	Kind            string  `comment:"Type of the operation."                                                                        json:"kind"`
	Signature       string  `comment:"Signature of the operation."                                                                   json:"signature"`
	Protocol        string  `comment:"Hash of the protocol, in which the operation was included in mempool."                         json:"protocol"`
	Level           uint64  `comment:"The height of the block from the genesis block, in which the operation was included."          json:"level"`
	Errors          JSONB   `bun:",type:jsonb"                                                                                       comment:"Errors with the operation processing if any." json:"errors,omitempty"`
	ExpirationLevel *uint64 `comment:"Datetime of block expiration in which the operation was included in seconds since UNIX epoch." json:"expiration_level"`
	Raw             JSONB   `bun:",type:jsonb"                                                                                       comment:"Raw JSON object of the operation."            json:"raw,omitempty"`
}

var _ bun.BeforeAppendModelHook = (*MempoolOperation)(nil)

func (mo *MempoolOperation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		mo.CreatedAt = time.Now().Unix()
	case *bun.UpdateQuery:
		mo.UpdatedAt = time.Now().Unix()
	}
	return nil
}

// SetInChain -
func SetInChain(ctx context.Context, db bun.IDB, network, hash, kind string, level uint64) error {
	model, err := getModelByKind(kind)
	if err != nil {
		return err
	}

	if _, err := db.NewUpdate().Model(model).
		Where("hash = ?", hash).
		Where("network = ?", network).
		Set("status = ?", StatusInChain).
		Set("level = ?", level).
		Set("errors = NULL").
		Exec(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

// SetExpired -
func SetExpired(ctx context.Context, db bun.IDB, network, branch string, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}

		if _, err := db.NewUpdate().
			Model(model).
			Set("status = ?", StatusExpired).
			Where("network = ?", network).
			Where("branch = ?", branch).
			Where("status = ?", StatusApplied).
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Rollback -
func Rollback(ctx context.Context, db bun.IDB, network, branch string, level uint64, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}

		query := db.NewUpdate().Model(model).
			Where("network = ?", network).
			Where("branch = ?", branch).
			Set("status = ?", StatusBranchRefused).
			WhereGroup(" AND ", func(q *bun.UpdateQuery) *bun.UpdateQuery {
				return q.Where("status = ?", StatusApplied).WhereGroup(" OR ", func(q1 *bun.UpdateQuery) *bun.UpdateQuery {
					return q1.Where("status = ?", StatusInChain).Where("level = ?", level)
				})
			})

		if _, err := query.Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewUpdate().Model(model).
			Set("status = ?", StatusApplied).
			Where("network = ?", network).
			Where("branch = ?", branch).
			Where("status = ?", StatusInChain).
			Where("level < ?", level).
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

// DeleteOldOperations -
func DeleteOldOperations(ctx context.Context, db bun.IDB, timeout uint64, status string, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}
		ts := time.Now().Unix() - int64(timeout)
		query := db.NewDelete().Model(model).Where("updated_at < ?", ts)

		if status != "" {
			query.Where("status = ?", status)
		}

		if _, err := query.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetModelsBy -
func GetModelsBy(kinds ...string) []interface{} {
	var hasManager bool
	data := make([]interface{}, 0, len(kinds))
	for i := range kinds {
		hasManager = hasManager || node.IsManager(kinds[i])
		model, err := getModelByKind(kinds[i])
		if err == nil {
			data = append(data, model)
		}
	}

	if hasManager {
		data = append(data, &GasStats{})
	}
	return data
}

func getModelByKind(kind string) (interface{}, error) {
	switch kind {
	case node.KindActivation:
		return &ActivateAccount{}, nil
	case node.KindBallot:
		return &Ballot{}, nil
	case node.KindDelegation:
		return &Delegation{}, nil
	case node.KindDoubleBaking:
		return &DoubleBaking{}, nil
	case node.KindDoubleEndorsing:
		return &DoubleEndorsing{}, nil
	case node.KindEndorsement:
		return &Endorsement{}, nil
	case node.KindNonceRevelation:
		return &NonceRevelation{}, nil
	case node.KindOrigination:
		return &Origination{}, nil
	case node.KindProposal:
		return &Proposal{}, nil
	case node.KindReveal:
		return &Reveal{}, nil
	case node.KindTransaction:
		return &Transaction{}, nil
	case node.KindRegisterGlobalConstant:
		return &RegisterGlobalConstant{}, nil
	case node.KindDoublePreendorsement:
		return &DoublePreendorsing{}, nil
	case node.KindPreendorsement:
		return &Preendorsement{}, nil
	case node.KindSetDepositsLimit:
		return &SetDepositsLimit{}, nil
	case node.KindTransferTicket:
		return &TransferTicket{}, nil
	case node.KindTxRollupCommit:
		return &TxRollupCommit{}, nil
	case node.KindTxRollupDispatchTickets:
		return &TxRollupDispatchTickets{}, nil
	case node.KindTxRollupFinalizeCommitment:
		return &TxRollupFinalizeCommitment{}, nil
	case node.KindTxRollupOrigination:
		return &TxRollupOrigination{}, nil
	case node.KindTxRollupRejection:
		return &TxRollupRejection{}, nil
	case node.KindTxRollupRemoveCommitment:
		return &TxRollupRemoveCommitment{}, nil
	case node.KindTxRollupReturnBond:
		return &TxRollupReturnBond{}, nil
	case node.KindTxRollupSubmitBatch:
		return &TxRollupSubmitBatch{}, nil
	case node.KindIncreasePaidStorage:
		return &IncreasePaidStorage{}, nil
	case node.KindVdfRevelation:
		return &VdfRevelation{}, nil
	case node.KindDrainDelegate:
		return &DelegateDrain{}, nil
	case node.KindUpdateConsensusKey:
		return &UpdateConsensusKey{}, nil
	case node.KindSrAddMessages:
		return &SmartRollupAddMessage{}, nil
	case node.KindSrCement:
		return &SmartRollupCement{}, nil
	case node.KindSrExecute:
		return &SmartRollupExecute{}, nil
	case node.KindSrOriginate:
		return &SmartRollupOriginate{}, nil
	case node.KindSrPublish:
		return &SmartRollupPublish{}, nil
	case node.KindSrRecoverBond:
		return &SmartRollupRecoverBond{}, nil
	case node.KindSrRefute:
		return &SmartRollupRefute{}, nil
	case node.KindSrTimeout:
		return &SmartRollupTimeout{}, nil
	case node.KindDalPublishCommitment:
		return &DalPublishCommitment{}, nil


	default:
		return nil, errors.Wrap(node.ErrUnknownKind, kind)
	}

}
