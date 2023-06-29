package models

// DoubleBaking -
type DoubleBaking struct {
	//nolint
	tableName struct{} `pg:"double_bakings" comment:"double_baking operation - is used by bakers to provide evidence of double baking (baking two different blocks at the same height) by a baker."`

	MempoolOperation
	Bh1 DoubleBakingInfo `json:"bh1" pg:"-"`
	Bh2 DoubleBakingInfo `json:"bh2" pg:"-"`

	Bh1Level            uint64 `json:"-" pg:"bh1_level" comment:"Height of the first block from the genesis."`
	Bh1Proto            int64  `json:"-" pg:"bh1_proto" comment:"First block protocol code, representing a number of protocol changes since genesis (mod 256, but -1 for the genesis block)."`
	Bh1ValidationPass   int64  `json:"-" pg:"bh1_validation_pass" comment:"First block number of endorsements (slots), included into the block."`
	Bh1Priority         int64  `json:"-" pg:"bh1_priority" comment:"First block priority [DEPRECATED]."`
	Bh1ProofOfWorkNonce string `json:"-" pg:"bh1_proof_of_work_nonce" comment:"First block proof of work nonce."`

	Bh2Level            uint64 `json:"-" pg:"bh2_level" comment:"Height of the second block from the genesis."`
	Bh2Proto            int64  `json:"-" pg:"bh2_proto" comment:"Second block protocol code, representing a number of protocol changes since genesis (mod 256, but -1 for the genesis block)."`
	Bh2ValidationPass   int64  `json:"-" pg:"bh2_validation_pass" comment:"Second block number of endorsements (slots), included into the block."`
	Bh2Priority         int64  `json:"-" pg:"bh2_priority" comment:"Second block priority [DEPRECATED]."`
	Bh2ProofOfWorkNonce string `json:"-" pg:"bh2_proof_of_work_nonce" comment:"Second block proof of work nonce."`
}

// DoubleBakingInfo -
type DoubleBakingInfo struct {
	Level            uint64   `json:"level"`
	Proto            int64    `json:"proto"`
	ValidationPass   int64    `json:"validation_pass"`
	Fitness          []string `json:"fitness"`
	Priority         int64    `json:"priority"`
	ProofOfWorkNonce string   `json:"proof_of_work_nonce"`
}

// Fill -
func (mdb *DoubleBaking) Fill() {
	mdb.Bh1Proto = mdb.Bh1.Proto
	mdb.Bh1ValidationPass = mdb.Bh1.ValidationPass
	mdb.Bh1Level = mdb.Bh1.Level
	mdb.Bh1Priority = mdb.Bh1.Priority
	mdb.Bh1ProofOfWorkNonce = mdb.Bh1.ProofOfWorkNonce

	mdb.Bh2Proto = mdb.Bh2.Proto
	mdb.Bh2ValidationPass = mdb.Bh2.ValidationPass
	mdb.Bh2Level = mdb.Bh2.Level
	mdb.Bh2Priority = mdb.Bh2.Priority
	mdb.Bh2ProofOfWorkNonce = mdb.Bh2.ProofOfWorkNonce
}
