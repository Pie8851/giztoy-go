package gizrun

import (
	"net/http"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	gizrunHTTPRequestsTotal          = "gizrun_http_requests_total"
	gizrunHTTPRequestDurationSeconds = "gizrun_http_request_duration_seconds"
)

func init() {
	InitAt(initSeqFlags, func() error {
		metrics.Reset(prometheus.DefaultRegisterer)
		return nil
	})
	ExitAt(exitSeqRuntime+1, func(*RunContext) error {
		metrics.Reset(prometheus.DefaultRegisterer)
		return nil
	})
	PostInitAt(postInitSeqRuntime, func(ctx *RunContext) error {
		ctx.serveMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			metrics.Handler().ServeHTTP(w, r)
		})
		if _, err := metrics.RegisterCounter(gizrunHTTPRequestsTotal,
			httpMethod,
			httpPath,
			httpHost,
			httpStatusCode,
		); err != nil {
			panic(err)
		}
		if _, err := metrics.RegisterHistogram(gizrunHTTPRequestDurationSeconds,
			httpMethod,
			httpPath,
			httpHost,
			httpStatusCode,
		); err != nil {
			panic(err)
		}
		return nil
	})
}
