package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "demo_app"
	subsystem = "web"
)

type Metrics struct {
	RequestsTotal  prometheus.Counter
	RequestLatency prometheus.Gauge
	Errors         *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	requestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Total number of requests processed",
	})
	requestLatency := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_latency_seconds",
		Help:      "Request latency in seconds",
	})
	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "errors_total",
		Help:      "Total number of errors by type",
	}, []string{"type"})

	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(errors)

	return &Metrics {
		RequestsTotal:requestsTotal,
		RequestLatency: requestLatency,
		Errors: errors,
	}
}
