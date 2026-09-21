package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-api-starter/internal/transport"
)

// Metrics contains the application HTTP metrics. A private registry keeps
// tests and multiple router instances isolated from the process global state.
type Metrics struct {
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
	Registry *prometheus.Registry
}

func New() *Metrics {
	registry := prometheus.NewRegistry()
	m := &Metrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "go_api_starter",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "go_api_starter",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
		}, []string{"method", "route"}),
		Registry: registry,
	}
	registry.MustRegister(m.Requests, m.Duration)
	return m
}

func (m *Metrics) Middleware() transport.HandlerFunc {
	return func(c *transport.Context) {
		started := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		method := c.Request.Method
		m.Requests.WithLabelValues(method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.Duration.WithLabelValues(method, route).Observe(time.Since(started).Seconds())
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
