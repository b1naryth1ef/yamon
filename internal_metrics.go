package yamon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ingestedMetrics = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "yamon_ingested_metrics",
		Help: "The number of ingested metrics",
	}, []string{"result"})

	ingestedLogs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "yamon_ingested_logs",
		Help: "The number of ingested logs",
	}, []string{"result"})
)
