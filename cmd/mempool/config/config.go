package config

import (
	"github.com/dipdup-net/go-lib/config"
)

// Config
type Config struct {
	config.Config `yaml:",inline"`
	Mempool       Mempool `yaml:"mempool" validate:"required"`
}

// Mempool -
type Mempool struct {
	Indexers map[string]*Indexer `yaml:"indexers" validate:"required"`
	Settings Settings            `yaml:"settings" validate:"required"`
}

// Indexer -
type Indexer struct {
	Filters    Filters           `yaml:"filters" validate:"required"`
	DataSource MempoolDataSource `yaml:"datasources" validate:"required"`
}

// Filters -
type Filters struct {
	Accounts []*config.Alias[config.Contract] `yaml:"accounts" validate:"max=50"`
	Kinds    []string                         `yaml:"kinds" validate:"required,min=1,dive,oneof=activate_account ballot delegation double_baking_evidence double_endorsement_evidence endorsement endorsement_with_slot origination proposals reveal seed_nonce_revelation transaction register_global_constant"`
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
	Tzkt *config.Alias[config.DataSource]   `yaml:"tzkt" validate:"required,url"`
	RPC  []*config.Alias[config.DataSource] `yaml:"rpc" validate:"required,min=1,dive,url"`
}

// URLs -
func (ds MempoolDataSource) URLs() []string {
	urls := make([]string, 0)
	for i := range ds.RPC {
		urls = append(urls, ds.RPC[i].Struct().URL)
	}
	return urls
}

// Settings -
type Settings struct {
	KeepOperations    uint64 `yaml:"keep_operations_seconds" validate:"required,min=1"`
	ExpiredAfter      uint64 `yaml:"expired_after_blocks" validate:"required,min=1"`
	KeepInChainBlocks uint64 `yaml:"keep_in_chain_blocks" validate:"required,min=1"`
	GasStatsLifetime  uint64 `yaml:"gas_stats_lifetime" validate:"required,min=1"`
}
