package observability

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var workerDurationBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10}

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
			Buckets: workerDurationBuckets,
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
	workerEventsTotal.WithLabelValues(normalizeResult(result)).Inc()
}

func normalizeResult(result string) string {
	result = strings.ToLower(strings.TrimSpace(result))
	if result == "" {
		return "unknown"
	}
	return result
}