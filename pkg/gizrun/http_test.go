package gizrun

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/labelset"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHTTPHandlerTagsRequestAndRecordsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics.Reset(reg)
	t.Cleanup(func() { metrics.Reset(prometheus.DefaultRegisterer) })
	if _, err := metrics.RegisterCounter(gizrunHTTPRequestsTotal,
		httpMethod,
		httpPath,
		httpHost,
		httpStatusCode,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := metrics.RegisterHistogram(gizrunHTTPRequestDurationSeconds,
		httpMethod,
		httpPath,
		httpHost,
		httpStatusCode,
	); err != nil {
		t.Fatal(err)
	}

	var labels labelset.LabelSet
	runCtx.serveMux = http.NewServeMux()
	runCtx.serveMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.Handler().ServeHTTP(w, r)
	})
	runCtx.serveMux.HandleFunc("/v1/test", func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		labels, ok = httpLabels(r.Context())
		if !ok {
			t.Fatal("httpLabels missing")
		}
		w.WriteHeader(http.StatusCreated)
	})
	runCtx.httpHandler = wrapHandler(runCtx.serveMux)

	req := httptest.NewRequest(http.MethodPost, "/v1/test", nil)
	req.Host = "gizclaw.test"
	rec := httptest.NewRecorder()
	runCtx.httpHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	for key, want := range map[string]string{
		httpMethod: http.MethodPost,
		httpPath:   "/v1/test",
		httpHost:   "gizclaw.test",
	} {
		if got, ok := labels.Value(key); !ok || got != want {
			t.Fatalf("HTTP label %q = (%q, %v), want (%q, true)", key, got, ok, want)
		}
	}
	assertGatheredMetric(t, reg, "gizclaw_gizrun_http_requests_total")
	assertGatheredMetric(t, reg, "gizclaw_gizrun_http_request_duration_seconds")
}

func assertGatheredMetric(t *testing.T, gatherer prometheus.Gatherer, name string) {
	t.Helper()
	families, err := gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() == name {
			return
		}
	}
	t.Fatalf("metric %q was not gathered", name)
}
