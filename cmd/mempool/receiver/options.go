package receiver

import (
	"github.com/dipdup-net/go-lib/prometheus"
)

// ReceiverOption -
type ReceiverOption func(*Receiver)

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
