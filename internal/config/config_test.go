package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		envs     map[string]string
		want     *Config
		wantErr  bool
	}{
		{
			name:     "config 1",
			filename: "./test/config1.yaml",
			want: &Config{
				Version: "0.0.1",
				DataSources: map[string]DataSource{
					"tzkt_mainnet": {
						Kind: "tzkt",
						URL:  "https://staging.api.tzkt.io",
					},
					"node_mainnet": {
						Kind: "tezos-node",
						URL:  "https://rpc.tzkt.io/mainnet",
					},
				},
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		}, {
			name:     "config 2",
			filename: "./test/config2.yaml",
			want: &Config{
				Version: "0.0.1",
				Contracts: map[string]Contract{
					"test": {
						Address: "KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9",
					},
				},
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		}, {
			name:     "config 3",
			filename: "./test/config3.yaml",
			want: &Config{
				Version: "0.0.1",
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		}, {
			name:     "config 4",
			filename: "./test/config4.yaml",
			want: &Config{
				Version: "0.0.1",
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Contracts: map[string]Contract{
					"test": {
						Address: "KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9",
					},
				},
				DataSources: map[string]DataSource{
					"tzkt_mainnet": {
						Kind: "tzkt",
						URL:  "https://staging.api.tzkt.io",
					},
					"node_mainnet": {
						Kind: "tezos-node",
						URL:  "https://rpc.tzkt.io/mainnet",
					},
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		}, {
			name:     "config 5 without envs",
			filename: "./test/config5.yaml",
			want: &Config{
				Version: "0.0.1",
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		}, {
			name:     "config 5 with envs",
			filename: "./test/config5.yaml",
			envs: map[string]string{
				"ACCOUNT": "test",
			},
			want: &Config{
				Version: "0.0.1",
				Database: Database{
					Kind: "sqlite",
					Path: "mempool.db",
				},
				Mempool: Mempool{
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"test"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://staging.api.tzkt.io",
								RPC:  []string{"https://rpc.tzkt.io/mainnet"},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for name := range tt.envs {
					os.Unsetenv(name)
				}
			}()

			for name, val := range tt.envs {
				os.Setenv(name, val)
			}

			got, err := Load(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
