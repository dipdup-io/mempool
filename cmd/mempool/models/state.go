package models

import (
	"fmt"
)

const (
	IndexTypeMempool = "mempool"
)

// MempoolIndexName -
func MempoolIndexName(network string) string {
	return fmt.Sprintf("%s_%s", IndexTypeMempool, network)
}
