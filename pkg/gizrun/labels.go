package gizrun

import (
	"context"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/labelset"
)

const (
	nsHTTP    = "http"
	nsGenx    = "genx"
	nsLogSink = "logsink"
)

const (
	httpMethod     = "method"
	httpRoute      = "route"
	httpPath       = "path"
	httpHost       = "host"
	httpStatusCode = "status_code"
)

const (
	genxProvider  = "provider"
	genxMethod    = "method"
	genxModel     = "model"
	genxStatus    = "status"
	genxTokenType = "token_type"
)

const (
	tokenCached    = "cached"
	tokenGenerated = "generated"
	tokenPrompt    = "prompt"
)

func tagHTTP(ctx context.Context, keyValues ...string) context.Context {
	return labelset.Tag(ctx, nsHTTP, keyValues...)
}

func tagGenx(ctx context.Context, keyValues ...string) context.Context {
	return labelset.Tag(ctx, nsGenx, keyValues...)
}

func tagLogSink(ctx context.Context, keyValues ...string) context.Context {
	return labelset.Tag(ctx, nsLogSink, keyValues...)
}

func tag(ctx context.Context, name string, keyValues ...string) context.Context {
	return labelset.Tag(ctx, name, keyValues...)
}

func httpLabels(ctx context.Context) (labelset.LabelSet, bool) {
	return labelset.FromContext(ctx, nsHTTP)
}

func genxLabels(ctx context.Context) (labelset.LabelSet, bool) {
	return labelset.FromContext(ctx, nsGenx)
}

func logSinkLabels(ctx context.Context) (labelset.LabelSet, bool) {
	return labelset.FromContext(ctx, nsLogSink)
}

func labels(ctx context.Context, namespace string) (labelset.LabelSet, bool) {
	return labelset.FromContext(ctx, namespace)
}
