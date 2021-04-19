package config

import (
	"net/url"

	"github.com/dipdup-net/mempool/internal/node"
	"github.com/pkg/errors"
)

// Validate -
func (c *Config) Validate() error {
	for network, mempool := range c.Mempool.Indexers {
		if err := mempool.DataSource.Validate(); err != nil {
			return errors.Wrap(err, network)
		}
		if err := mempool.Filters.Validate(); err != nil {
			return errors.Wrap(err, network)
		}
	}
	return c.Mempool.Settings.Validate()
}

// Validate -
func (cfg MempoolDataSource) Validate() error {
	if _, err := url.Parse(cfg.Tzkt); err != nil {
		return errors.Wrapf(err, "Invalid TzKT url %s", cfg.Tzkt)
	}
	if len(cfg.RPC) == 0 {
		return errors.Errorf("Empty nodes list")
	}

	for i := range cfg.RPC {
		if _, err := url.Parse(cfg.RPC[i]); err != nil {
			return errors.Wrapf(err, "Invalid TzKT url %s", cfg.RPC[i])
		}
	}
	return nil
}

// Validate -
func (cfg Filters) Validate() error {
	switch {
	case len(cfg.Kinds) == 0:
		cfg.Kinds = append(cfg.Kinds, node.KindTransaction)
	default:
		if err := validateKinds(cfg.Kinds...); err != nil {
			return err
		}
	}

	if len(cfg.Accounts) > tzktMaxSubscriptions {
		return errors.Errorf("Maximum accounts in config is %d. You added %d accounts", tzktMaxSubscriptions, len(cfg.Accounts))
	}

	return nil
}

func validateKinds(kinds ...string) error {
	for _, kind := range kinds {
		var found bool

		for _, valid := range []string{
			node.KindActivation, node.KindBallot, node.KindDelegation, node.KindDoubleBaking, node.KindDoubleEndorsing,
			node.KindEndorsement, node.KindNonceRevelation, node.KindOrigination, node.KindProposal,
			node.KindReveal, node.KindTransaction,
		} {
			if kind == valid {
				found = true
				break
			}
		}

		if !found {
			return errors.Wrap(node.ErrUnknownKind, kind)
		}
	}
	return nil
}

// Validate -
func (settings *Settings) Validate() error {
	if settings.ExpiredAfter == 0 {
		settings.ExpiredAfter = 60
	}
	if settings.KeepInChainBlocks == 0 {
		settings.KeepInChainBlocks = 10
	}
	if settings.KeepOperations == 0 {
		settings.KeepOperations = 172800
	}
	if settings.MempoolRequestInterval == 0 {
		settings.MempoolRequestInterval = 10
	}
	if settings.RPCTimeout == 0 {
		settings.RPCTimeout = 10
	}
	return nil
}
