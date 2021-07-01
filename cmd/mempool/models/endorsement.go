package models

import (
	"bytes"

	"gorm.io/gorm"
)

// Endorsement -
type Endorsement struct {
	MempoolOperation
	Level uint64 `json:"level"`
	Baker string `json:"-" gorm:"transaction_baker_idx"`
}

// TableName -
func (Endorsement) TableName() string {
	return "endorsements"
}

// Forge -
func (endorsement Endorsement) Forge() ([]byte, error) {
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
func EndorsementsWithoutBaker(tx *gorm.DB) (endorsements []Endorsement, err error) {
	err = tx.Model(&Endorsement{}).Where("baker = ''").Order("level asc").Find(&endorsements).Error
	return
}
