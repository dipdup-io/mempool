package models

// MempoolDoubleBaking -
type MempoolDoubleBaking struct {
	MempoolOperation
	Bh1 DoubleBakingInfo `json:"bh1" gorm:"-"`
	Bh2 DoubleBakingInfo `json:"bh2" gorm:"-"`

	Bh1Level            uint64 `json:"-"`
	Bh1Proto            int64  `json:"-"`
	Bh1ValidationPass   int64  `json:"-"`
	Bh1Priority         int64  `json:"-"`
	Bh1ProofOfWorkNonce string `json:"-"`

	Bh2Level            uint64 `json:"-"`
	Bh2Proto            int64  `json:"-"`
	Bh2ValidationPass   int64  `json:"-"`
	Bh2Priority         int64  `json:"-"`
	Bh2ProofOfWorkNonce string `json:"-"`
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
func (mdb *MempoolDoubleBaking) Fill() {
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
