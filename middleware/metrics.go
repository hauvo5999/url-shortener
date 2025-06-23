package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	urlsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "urls_created_total",
			Help: "Total number of URLs created",
		},
	)

	urlRedirectsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "url_redirects_total", 
			Help: "Total number of URL redirects",
		},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		endpoint := r.URL.Path
		
		httpRequestsTotal.WithLabelValues(
			r.Method,
			endpoint,
			strconv.Itoa(rw.statusCode),
		).Inc()
		
		httpRequestDuration.WithLabelValues(
			r.Method,
			endpoint,
		).Observe(duration.Seconds())
		
		// Track specific metrics
		if r.Method == "POST" && endpoint == "/shorten" && rw.statusCode == http.StatusOK {
			urlsCreatedTotal.Inc()
		}
		
		if r.Method == "GET" && endpoint != "/health" && endpoint != "/metrics" && rw.statusCode == http.StatusMovedPermanently {
			urlRedirectsTotal.Inc()
		}
	})
}