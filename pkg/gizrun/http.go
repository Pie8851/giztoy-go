package gizrun

import (
	"net/http"
	"strconv"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/metrics"
)

func wrapHandler(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := tagHTTP(r.Context(),
			httpMethod, r.Method,
			httpPath, r.URL.Path,
			httpHost, r.Host,
		)
		recorder := &statusResponseWriter{ResponseWriter: w}
		inner.ServeHTTP(recorder, r.WithContext(ctx))
		statusCode := recorder.statusCode()
		metricCtx := tagHTTP(ctx, httpStatusCode, strconv.Itoa(statusCode))
		if labels, ok := httpLabels(metricCtx); ok {
			counter := metrics.Counter(gizrunHTTPRequestsTotal)
			histogram := metrics.Histogram(gizrunHTTPRequestDurationSeconds)
			if counter != nil && histogram != nil {
				promLabels := labels.PrometheusLabels()
				counter.With(promLabels).Inc()
				histogram.With(promLabels).Observe(time.Since(start).Seconds())
			}
		}
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusResponseWriter) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *statusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
