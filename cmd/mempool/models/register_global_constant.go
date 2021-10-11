package models

import "gorm.io/datatypes"

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	MempoolOperation
	Source       string         `json:"source"`
	Fee          string         `json:"fee"`
	Counter      string         `json:"counter"`
	GasLimit     string         `json:"gas_limit"`
	StorageLimit string         `json:"storage_limit"`
	Value        datatypes.JSON `json:"value"`
}

// TableName -
func (RegisterGlobalConstant) TableName() string {
	return "register_global_constant"
}
