package receiver

import "gorm.io/gorm"

// ReceiverOption -
type ReceiverOption func(*Receiver)

// WithInterval -
func WithInterval(seconds uint64) ReceiverOption {
	return func(m *Receiver) {
		if seconds > 0 {
			m.interval = seconds
		} else {
			m.interval = 10
		}
	}
}

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
