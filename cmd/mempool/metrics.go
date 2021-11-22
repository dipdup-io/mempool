package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	operationCountMetricName = "mempool_operation_count"
	rpcErrorsCountName       = "mempool_rpc_errors_count"
)

var (
	counters = make(map[string]*prometheus.CounterVec)
)

func registerPrometheusMetrics() {
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	operationsCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: operationCountMetricName,
		Help: "The total number operations in mempool DipDup",
	}, []string{"kind", "status", "network"})
	prometheus.MustRegister(operationsCount)
	counters[operationCountMetricName] = operationsCount

	rpcErrorsCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: rpcErrorsCountName,
		Help: "The total number of RPC errors in mempool DipDup",
	}, []string{"code", "node", "network"})
	prometheus.MustRegister(rpcErrorsCount)
	counters[rpcErrorsCountName] = rpcErrorsCount
}

func incrementMetric(name string, labels map[string]string) {
	counter, ok := counters[name]
	if ok {
		counter.With(labels).Inc()
	}
}

func getMetric(name string) *prometheus.CounterVec {
	counter, ok := counters[name]
	if ok {
		return counter
	}
	return nil
}
