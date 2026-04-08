package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	workerEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finpharm_worker_events_total",
			Help: "Total worker events processed by result.",
		},
		[]string{"result"},
	)

	workerProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "finpharm_worker_processing_duration_seconds",
			Help:    "Duration of worker message processing.",
			Buckets: prometheus.DefBuckets,
		},
	)

	workerInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "finpharm_worker_inflight_messages",
			Help: "Current number of in-flight worker messages.",
		},
	)
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func BeginProcessing() func() {
	workerInFlight.Inc()
	start := time.Now()

	return func() {
		workerInFlight.Dec()
		workerProcessingDuration.Observe(time.Since(start).Seconds())
	}
}

func IncResult(result string) {
	workerEventsTotal.WithLabelValues(result).Inc()
}