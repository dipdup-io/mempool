package models

import (
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

// EndorsementsWithoutBaker -
func EndorsementsWithoutBaker(tx *gorm.DB) (endorsements []Endorsement, err error) {
	err = tx.Model(&Endorsement{}).Where("baker = ''").Order("level asc").Find(&endorsements).Error
	return
}
