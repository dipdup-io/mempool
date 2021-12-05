package config

import (
	"os"
	"testing"

	"github.com/dipdup-net/go-lib/config"
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
				Config: config.Config{
					Version: "0.0.1",
					DataSources: map[string]config.DataSource{
						"tzkt_mainnet": {
							Kind: "tzkt",
							URL:  "https://api.tzkt.io",
						},
						"node_mainnet": {
							Kind: "tezos-node",
							URL:  "https://mainnet-tezos.giganode.io",
						},
					},
					Database: config.Database{
						Kind: "sqlite",
						Path: "mempool.db",
					},
					Hasura: &config.Hasura{
						URL:                "http://hasura:8080",
						Secret:             "admin_secret",
						RowsLimit:          100,
						EnableAggregations: true,
					},

					Prometheus: &config.Prometheus{
						URL: "127.0.0.1:2112",
					},
				},
				Mempool: Mempool{
					Settings: Settings{
						KeepOperations:    172800,
						ExpiredAfter:      120,
						KeepInChainBlocks: 10,
						GasStatsLifetime:  3600,
					},
					Indexers: map[string]*Indexer{
						"mainnet": {
							Filters: Filters{
								Kinds:    []string{"transaction"},
								Accounts: []string{"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://api.tzkt.io",
								RPC:  []string{"https://mainnet-tezos.giganode.io"},
							},
						},
						"granadanet": {
							Filters: Filters{
								Kinds: []string{"endorsement"},
							},
							DataSource: MempoolDataSource{
								Tzkt: "https://api.granadanet.tzkt.io",
								RPC:  []string{"https://testnet-tezos.giganode.io"},
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

			var got Config
			err := config.Parse(tt.filename, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, &got, tt.want)
		})
	}
}
