package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
			Buckets: prometheus.DefBuckets,
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
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func InstrumentHandler(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		httpInFlightRequests.WithLabelValues(service).Inc()
		defer httpInFlightRequests.WithLabelValues(service).Dec()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		status := strconv.Itoa(rec.status)
		path := r.URL.Path

		httpRequestsTotal.WithLabelValues(service, r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(service, r.Method, path, status).Observe(time.Since(start).Seconds())
	})
}