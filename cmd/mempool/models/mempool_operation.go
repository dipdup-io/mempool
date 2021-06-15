package models

import (
	"time"

	"github.com/dipdup-net/go-lib/node"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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

// MempoolOperation -
type MempoolOperation struct {
	UpdatedAt       int
	Network         string  `gorm:"primaryKey" json:"network"`
	Hash            string  `gorm:"primaryKey" json:"hash"`
	Branch          string  `json:"branch"`
	Status          string  `json:"status"`
	Kind            string  `json:"kind"`
	Signature       string  `json:"signature"`
	Level           uint64  `json:"level"`
	Errors          JSON    `json:"errors,omitempty"`
	ExpirationLevel *uint64 `json:"expiration_level"`
}

func networkAndBranch(network, branch string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).Where("branch = ?", branch)
	}
}

func isApplied(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", StatusApplied)
}

func isInChain(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", StatusInChain)
}

// SetInChain -
func SetInChain(db *gorm.DB, network, hash, kind string, level uint64) error {
	model, err := getModelByKind(kind)
	if err != nil {
		return err
	}
	query := db.Model(model).Where("network = ? AND hash = ?", network, hash)

	if err := query.Updates(map[string]interface{}{
		"status": StatusInChain,
		"level":  level,
		"errors": nil,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// SetExpired -
func SetExpired(db *gorm.DB, network, branch string, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}
		if err := db.Model(model).
			Scopes(networkAndBranch(network, branch), isApplied).
			Update("status", StatusExpired).Error; err != nil {
			return err
		}
	}
	return nil
}

// Rollback -
func Rollback(db *gorm.DB, network, branch string, level uint64, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}
		if err := db.Model(model).
			Scopes(networkAndBranch(network, branch)).
			Where(
				db.Where("status = ?", StatusApplied).
					Or(
						db.Where("status = ?", StatusInChain).Where("level = ?", level),
					),
			).
			Update("status", StatusBranchRefused).Error; err != nil {
			return err
		}

		if err := db.Model(model).
			Scopes(networkAndBranch(network, branch), isInChain).
			Where("level < ?", level).
			Update("status", StatusApplied).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteOldOperations -
func DeleteOldOperations(db *gorm.DB, timeout uint64, status string, kinds ...string) error {
	if len(kinds) == 0 {
		return nil
	}

	for _, kind := range kinds {
		model, err := getModelByKind(kind)
		if err != nil {
			return err
		}
		ts := time.Now().Unix() - int64(timeout)
		query := db.Where("updated_at < ?", ts)

		if status != "" {
			query.Where("status = ?", status)
		}

		if err := query.Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetModelsBy -
func GetModelsBy(kinds ...string) []interface{} {
	data := make([]interface{}, 0, len(kinds))
	for i := range kinds {
		model, err := getModelByKind(kinds[i])
		if err == nil {
			data = append(data, model)
		}
	}
	return data
}

func getModelByKind(kind string) (interface{}, error) {
	switch kind {
	case node.KindActivation:
		return &MempoolActivateAccount{}, nil
	case node.KindBallot:
		return &MempoolBallot{}, nil
	case node.KindDelegation:
		return &MempoolDelegation{}, nil
	case node.KindDoubleBaking:
		return &MempoolDoubleBaking{}, nil
	case node.KindDoubleEndorsing:
		return &MempoolDoubleEndorsing{}, nil
	case node.KindEndorsement:
		return &MempoolEndorsement{}, nil
	case node.KindNonceRevelation:
		return &MempoolNonceRevelation{}, nil
	case node.KindOrigination:
		return &MempoolOrigination{}, nil
	case node.KindProposal:
		return &MempoolProposal{}, nil
	case node.KindReveal:
		return &MempoolReveal{}, nil
	case node.KindTransaction:
		return &MempoolTransaction{}, nil
	default:
		return nil, errors.Wrap(node.ErrUnknownKind, kind)
	}

}
