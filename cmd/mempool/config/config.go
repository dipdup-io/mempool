package config

import (
	"github.com/dipdup-net/go-lib/config"
)

// Config
type Config struct {
	config.Config `yaml:",inline"`
	Mempool       Mempool `yaml:"mempool"`
}

// Mempool -
type Mempool struct {
	Indexers map[string]*Indexer `yaml:"indexers"`
	Settings Settings            `yaml:"settings"`
}

// Indexer -
type Indexer struct {
	Filters    Filters           `yaml:"filters"`
	DataSource MempoolDataSource `yaml:"datasources"`
}

// Filters -
type Filters struct {
	Accounts []string `yaml:"accounts"`
	Kinds    []string `yaml:"kinds"`
}

// MempoolDataSource -
type MempoolDataSource struct {
	Tzkt string   `yaml:"tzkt"`
	RPC  []string `yaml:"rpc"`
}

// Settings -
type Settings struct {
	KeepOperations         uint64 `yaml:"keep_operations_seconds"`
	ExpiredAfter           uint64 `yaml:"expired_after_blocks"`
	KeepInChainBlocks      uint64 `yaml:"keep_in_chain_blocks"`
	MempoolRequestInterval uint64 `yaml:"mempool_request_interval_seconds"`
	RPCTimeout             uint64 `yaml:"rpc_timeout_seconds"`
}

// Load -
func Load(filename string) (cfg Config, err error) {
	err = config.Parse(filename, &cfg)
	return
}
