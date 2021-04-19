package models

// MempoolDoubleEndorsing -
type MempoolDoubleEndorsing struct {
	MempoolOperation
	Op1Kind  string `json:"-"`
	Op1Level uint64 `json:"-"`
	Op2Kind  string `json:"-"`
	Op2Level uint64 `json:"-"`

	Op1 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op1" gorm:"-"`
	Op2 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op2" gorm:"-"`
}

// DoubleEndorsingOperations -
type DoubleEndorsingOperations struct {
	Kind  string `json:"kind"`
	Level uint64 `json:"level"`
}

// Fill -
func (mde *MempoolDoubleEndorsing) Fill() {
	mde.Op1Kind = mde.Op1.Operations.Kind
	mde.Op1Level = mde.Op1.Operations.Level
	mde.Op2Kind = mde.Op2.Operations.Kind
	mde.Op2Level = mde.Op2.Operations.Level
}
