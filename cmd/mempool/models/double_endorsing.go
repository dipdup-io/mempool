package models

// DoubleEndorsing -
type DoubleEndorsing struct {
	//nolint
	tableName struct{} `pg:"double_endorsings"`

	MempoolOperation
	Op1Kind  string `json:"-" pg:"op1_kind"`
	Op1Level uint64 `json:"-" pg:"op1_level"`
	Op2Kind  string `json:"-" pg:"op2_kind"`
	Op2Level uint64 `json:"-" pg:"op2_level"`

	Op1 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op1" pg:"-"`
	Op2 struct {
		Operations DoubleEndorsingOperations `json:"operations"`
	} `json:"op2" pg:"-"`
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
