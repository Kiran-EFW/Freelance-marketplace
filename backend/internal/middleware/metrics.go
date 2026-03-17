package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed.",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsInFlight)
}

// NewPrometheusMiddleware returns a Fiber middleware that records Prometheus
// metrics for every request: total count, duration histogram, and in-flight gauge.
func NewPrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip the /metrics endpoint itself to avoid recursive counting.
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()
		method := c.Method()
		path := c.Path()

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		err := c.Next()

		status := strconv.Itoa(c.Response().StatusCode())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

// MetricsHandler returns a Fiber handler that serves the Prometheus metrics
// endpoint using promhttp.Handler() adapted for Fiber.
func MetricsHandler() fiber.Handler {
	// Adapt the standard net/http promhttp.Handler to a Fiber handler.
	return adaptor.HTTPHandler(promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
}

// metricsHTTPHandler is an alternative that returns a plain http.Handler
// (unused but available for testing / standalone use).
func metricsHTTPHandler() http.Handler {
	return promhttp.Handler()
}
