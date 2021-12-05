package receiver

import (
	"github.com/dipdup-net/go-lib/prometheus"
	pg "github.com/go-pg/pg/v10"
)

// ReceiverOption -
type ReceiverOption func(*Receiver)

// WithStorage -
func WithStorage(db *pg.DB, blockTime int64) ReceiverOption {
	return func(m *Receiver) {
		if blockTime > 0 {
			m.blockTime = blockTime
		} else {
			m.blockTime = 60
		}
		m.db = db
	}
}

// WithPrometheus -
func WithPrometheus(prom *prometheus.Service) ReceiverOption {
	return func(m *Receiver) {
		m.prom = prom
	}
}
