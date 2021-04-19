package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config
type Config struct {
	Version     string                `yaml:"version"`
	Database    Database              `yaml:"database"`
	DataSources map[string]DataSource `yaml:"datasources"`
	Contracts   map[string]Contract   `yaml:"contracts"`
	Mempool     Mempool               `yaml:"mempool"`
}

// DataSource -
type DataSource struct {
	Kind string `yaml:"kind"`
	URL  string `yaml:"url"`
}

// Contracts -
type Contract struct {
	Address  string `yaml:"address"`
	TypeName string `yaml:"typename"`
}

// Indexer -
type Indexer struct {
	Filters    Filters           `yaml:"filters"`
	DataSource MempoolDataSource `yaml:"datasources"`
}

// Mempool -
type Mempool struct {
	Indexers map[string]*Indexer `yaml:"indexers"`
	Settings Settings            `yaml:"settings"`
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
	LostAfter              uint64 `yaml:"lost_after_blocks"`
	KeepInChainBlocks      uint64 `yaml:"keep_in_chain_blocks"`
	MempoolRequestInterval uint64 `yaml:"mempool_request_interval_seconds"`
	RPCTimeout             uint64 `yaml:"rpc_timeout_seconds"`
}

// Database
type Database struct {
	Path string `yaml:"path"`
	Kind string `yaml:"kind"`
}

// Load - load config from `filename`
func Load(filename string) (*Config, error) {
	if filename == "" {
		return nil, fmt.Errorf("you have to provide configuration filename")
	}

	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s error: %w", filename, err)
	}

	expanded := expandEnv(string(src))

	var c Config
	if err := yaml.Unmarshal([]byte(expanded), &c); err != nil {
		return nil, err
	}

	return &c, c.substitute()
}

// Validate - validates config
func Validate(cfg *Config) error {
	if err := validateDBKind(cfg.Database.Kind); err != nil {
		return err
	}

	for network, mempool := range cfg.Mempool.Indexers {
		if err := validateDataSource(mempool.DataSource); err != nil {
			return errors.Wrap(err, network)
		}
		if err := validateFilters(mempool.Filters); err != nil {
			return errors.Wrap(err, network)
		}
	}
	return nil
}

// LoadAndValidate - load config from `filename` and validate it
func LoadAndValidate(filename string) (*Config, error) {
	cfg, err := Load(filename)
	if err != nil {
		return nil, err
	}
	return cfg, Validate(cfg)
}
