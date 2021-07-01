package models

// DoubleEndorsing -
type DoubleEndorsing struct {
	MempoolOperation
	Op1Kind  string `json:"-" gorm:"column:op_1_kind"`
	Op1Level uint64 `json:"-" gorm:"column:op_1_level"`
	Op2Kind  string `json:"-" gorm:"column:op_2_kind"`
	Op2Level uint64 `json:"-" gorm:"column:op_2_level"`

	Op1 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op1" gorm:"-"`
	Op2 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op2" gorm:"-"`
}

// TableName -
func (DoubleEndorsing) TableName() string {
	return "double_endorsings"
}

// DoubleEndorsingOperations -
type DoubleEndorsingOperations struct {
	Kind  string `json:"kind"`
	Level uint64 `json:"level"`
}

// Fill -
func (mde *DoubleEndorsing) Fill() {
	mde.Op1Kind = mde.Op1.Operations.Kind
	mde.Op1Level = mde.Op1.Operations.Level
	mde.Op2Kind = mde.Op2.Operations.Kind
	mde.Op2Level = mde.Op2.Operations.Level
}
