package receiver

import (
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/go-lib/prometheus"
)

// ReceiverOption -
type ReceiverOption func(*Receiver)

// WithStorage -
func WithStorage(db *database.PgGo) ReceiverOption {
	return func(m *Receiver) {
		m.db = db
	}
}

// WithPrometheus -
func WithPrometheus(prom *prometheus.Service) ReceiverOption {
	return func(m *Receiver) {
		m.prom = prom
	}
}

// WithBlockTime -
func WithBlockTime(blockTime int64) ReceiverOption {
	return func(m *Receiver) {
		m.blockTime = blockTime
	}
}
