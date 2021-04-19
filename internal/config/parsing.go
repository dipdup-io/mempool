package config

import (
	"os"
	"regexp"

	"github.com/pkg/errors"
)

var defaultEnv = regexp.MustCompile(`\${(?P<name>[\w\.]{1,}):-(?P<value>[\w\.:/]*)}`)

func expandEnv(data string) string {
	vars := defaultEnv.FindAllStringSubmatch(data, -1)
	data = defaultEnv.ReplaceAllString(data, `${$name}`)

	for i := range vars {
		if _, ok := os.LookupEnv(vars[i][1]); !ok {
			os.Setenv(vars[i][1], vars[i][2])
		}
	}

	data = os.ExpandEnv(data)
	return data
}

func (c *Config) substitute() error {
	for _, indexer := range c.Mempool.Indexers {
		if err := substituteContracts(c, &indexer.Filters); err != nil {
			return err
		}
		if err := substituteDataSources(c, &indexer.DataSource); err != nil {
			return err
		}
	}
	return nil
}

func substituteContracts(c *Config, filters *Filters) error {
	for i, address := range filters.Accounts {
		contract, ok := c.Contracts[address]
		if !ok {
			continue
		}
		filters.Accounts[i] = contract.Address
	}
	return nil
}

func substituteDataSources(c *Config, dataSource *MempoolDataSource) error {
	if source, ok := c.DataSources[dataSource.Tzkt]; ok {
		if source.Kind != DataSourceKindTzKT {
			return errors.Errorf("Invalid tzkt data source kind. Expected `tzkt`, got `%s`", source.Kind)
		}
		dataSource.Tzkt = source.URL
	}

	for i, link := range dataSource.RPC {
		source, ok := c.DataSources[link]
		if !ok {
			continue
		}
		if source.Kind != DataSourceKindNode {
			return errors.Errorf("Invalid RPC data source kind. Expected `tezos-node`, got `%s`", source.Kind)
		}
		dataSource.RPC[i] = source.URL
	}
	return nil
}
