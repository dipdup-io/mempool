package config

import (
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/mempool/cmd/mempool/profiler"
)

// Config
type Config struct {
	config.Config `yaml:",inline"`
	Mempool       Mempool          `validate:"required"       yaml:"mempool"`
	Profiler      *profiler.Config `yaml:"profiler,omitempty"`
}

// Mempool -
type Mempool struct {
	Indexers map[string]*Indexer `validate:"required" yaml:"indexers"`
	Settings Settings            `validate:"required" yaml:"settings"`
}

// Indexer -
type Indexer struct {
	Filters    Filters           `validate:"required" yaml:"filters"`
	DataSource MempoolDataSource `validate:"required" yaml:"datasources"`
}

// Filters -
type Filters struct {
	Accounts []*config.Alias[config.Contract] `validate:"max=50"                                                                                                                                                                                                                                    yaml:"accounts"`
	Kinds    []string                         `validate:"required,min=1,dive,oneof=activate_account ballot delegation double_baking_evidence double_endorsement_evidence endorsement endorsement_with_dal endorsement_with_slot origination proposals reveal seed_nonce_revelation transaction register_global_constant" yaml:"kinds"`
}

// Addresses -
func (f Filters) Addresses() []string {
	addresses := make([]string, 0)
	for i := range f.Accounts {
		addresses = append(addresses, f.Accounts[i].Struct().Address)
	}
	return addresses
}

// MempoolDataSource -
type MempoolDataSource struct {
	Tzkt *config.Alias[config.DataSource] `validate:"required,url"            yaml:"tzkt"`
	RPC  *config.Alias[config.DataSource] `validate:"required,min=1,dive,url" yaml:"rpc"`
}

// URLs -
func (ds MempoolDataSource) URL() string {
	if ds.RPC == nil {
		return ""
	}
	return ds.RPC.Struct().URL
}

// Settings -
type Settings struct {
	KeepOperations    uint64 `validate:"required,min=1" yaml:"keep_operations_seconds"`
	ExpiredAfter      uint64 `validate:"required,min=1" yaml:"expired_after_blocks"`
	KeepInChainBlocks uint64 `validate:"required,min=1" yaml:"keep_in_chain_blocks"`
	GasStatsLifetime  uint64 `validate:"required,min=1" yaml:"gas_stats_lifetime"`
}
