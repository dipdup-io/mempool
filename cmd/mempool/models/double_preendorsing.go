package models

// DoublePreendorsing -
type DoublePreendorsing struct {
	//nolint
	tableName struct{} `pg:"double_preendorsings"`

	MempoolOperation
	Op1Kind  string `json:"-" pg:"op_1_kind"`
	Op1Level uint64 `json:"-" pg:"op_1_level"`
	Op2Kind  string `json:"-" pg:"op_2_kind"`
	Op2Level uint64 `json:"-" pg:"op_2_level"`

	Op1 struct {
		Operations DoublePreendorsingOperations `json:"operations"`
	} `json:"op1" pg:"-"`
	Op2 struct {
		Operations DoublePreendorsingOperations `json:"operations"`
	} `json:"op2" pg:"-"`
}

// DoublePreendorsingOperations -
type DoublePreendorsingOperations struct {
	Kind  string `json:"kind"`
	Level uint64 `json:"level"`
}

// Fill -
func (mde *DoublePreendorsing) Fill() {
	mde.Op1Kind = mde.Op1.Operations.Kind
	mde.Op1Level = mde.Op1.Operations.Level
	mde.Op2Kind = mde.Op2.Operations.Kind
	mde.Op2Level = mde.Op2.Operations.Level
}
