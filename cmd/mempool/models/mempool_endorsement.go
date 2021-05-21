package models

import (
	"bytes"

	"gorm.io/gorm"
)

// MempoolEndorsement -
type MempoolEndorsement struct {
	MempoolOperation
	Level uint64 `json:"level"`
	Baker string `json:"-"`
}

// TableName -
func (MempoolEndorsement) TableName() string {
	return "mempool_endorsement"
}

// Forge -
func (endorsement MempoolEndorsement) Forge() ([]byte, error) {
	var buf bytes.Buffer

	if _, err := buf.Write(forgeString(endorsement.Branch)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(forgeNat(0)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(forgeInt(int64(endorsement.Level))); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// EndorsementsWithoutBaker -
func EndorsementsWithoutBaker(tx *gorm.DB) (endorsements []MempoolEndorsement, err error) {
	err = tx.Model(&MempoolEndorsement{}).Where("baker = ''").Order("level asc").Find(&endorsements).Error
	return
}
