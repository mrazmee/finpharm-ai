package observability

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpDurationBuckets = []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5}

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finpharm_http_requests_total",
			Help: "Total HTTP requests handled by the service.",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "finpharm_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: httpDurationBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)

	httpInFlightRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "finpharm_http_inflight_requests",
			Help: "Current number of in-flight HTTP requests.",
		},
		[]string{"service"},
	)

	transactionOutcomesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finpharm_transaction_outcomes_total",
			Help: "Total transaction outcomes by final status.",
		},
		[]string{"status"},
	)

	transactionAuditDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finpharm_transaction_audit_decisions_total",
			Help: "Total transaction audit decisions by decision/provider/model.",
		},
		[]string{"decision", "provider", "model"},
	)

	transactionReplaysTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "finpharm_transaction_replays_total",
			Help: "Total replayed create transaction requests due to idempotency key reuse.",
		},
	)
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func Middleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		httpInFlightRequests.WithLabelValues(service).Inc()
		defer httpInFlightRequests.WithLabelValues(service).Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := normalizePathLabel(c.FullPath())

		httpRequestsTotal.WithLabelValues(service, c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(service, c.Request.Method, path, status).Observe(time.Since(start).Seconds())
	}
}

func ObserveTransactionOutcome(status string) {
	transactionOutcomesTotal.WithLabelValues(normalizeLabel(status, "UNKNOWN")).Inc()
}

func ObserveTransactionAuditDecision(decision, provider, model string) {
	transactionAuditDecisionsTotal.WithLabelValues(
		normalizeLabel(decision, "UNKNOWN"),
		normalizeLabel(provider, "unknown"),
		normalizeLabel(model, "unknown"),
	).Inc()
}

func IncTransactionReplay() {
	transactionReplaysTotal.Inc()
}

func normalizeLabel(v string, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return strings.ToUpper(v)
}

func normalizePathLabel(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "unmatched"
	}
	return path
}