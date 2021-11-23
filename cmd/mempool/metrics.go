package main

import "github.com/dipdup-net/go-lib/prometheus"

const (
	operationCountMetricName = "mempool_operation_count"
	rpcErrorsCountName       = "mempool_rpc_errors_count"
)

func registerPrometheusMetrics(service *prometheus.Service) {
	// Add Go module build info.
	service.RegisterGoBuildMetrics()

	service.RegisterCounter(operationCountMetricName, "The total number operations in mempool DipDup", "kind", "status", "network")
	service.RegisterCounter(rpcErrorsCountName, "The total number of RPC errors in mempool DipDup", "code", "node", "network")

}
