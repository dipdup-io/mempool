package receiver

import (
	"github.com/dipdup-net/go-lib/prometheus"
	"gorm.io/gorm"
)

// ReceiverOption -
type ReceiverOption func(*Receiver)

// WithTimeout -
func WithTimeout(seconds uint64) ReceiverOption {
	return func(m *Receiver) {
		if seconds > 0 {
			m.timeout = seconds
		} else {
			m.timeout = 10
		}
	}
}

// WithStorage -
func WithStorage(db *gorm.DB, blockTime int64) ReceiverOption {
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
