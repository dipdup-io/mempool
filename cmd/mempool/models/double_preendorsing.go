package models

import "github.com/uptrace/bun"

// DoublePreendorsing -
type DoublePreendorsing struct {
	bun.BaseModel `bun:"double_preendorsings"`

	MempoolOperation
	Op1Kind  string `bun:"op1_kind"  comment:"Kind of the first operation."                                                            json:"-"`
	Op1Level uint64 `bun:"op1_level" comment:"Height of the block from the genesis block, in which the first operation was included."  json:"-"`
	Op2Kind  string `bun:"op2_kind"  comment:"Kind of the second operation."                                                           json:"-"`
	Op2Level uint64 `bun:"op2_level" comment:"Height of the block from the genesis block, in which the second operation was included." json:"-"`

	Op1 struct {
		Operations DoublePreendorsingOperations `json:"operations"`
	} `json:"op1" bun:"-"`
	Op2 struct {
		Operations DoublePreendorsingOperations `json:"operations"`
	} `json:"op2" bun:"-"`
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

// SetMempoolOperation -
func (mde *DoublePreendorsing) SetMempoolOperation(operaiton MempoolOperation) {
	mde.MempoolOperation = operaiton
}
