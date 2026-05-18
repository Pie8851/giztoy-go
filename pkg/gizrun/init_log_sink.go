package gizrun

import (
	"context"
	"log/slog"
	"os"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/log"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/log/sink"
)

func init() {
	PostInitAt(postInitSeqRuntime, func(ctx *RunContext) error {
		ctx.logHandlers = append(ctx.logHandlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &ctx.logLevel}))
		return nil
	})
	PostInitAt(postInitSeqRuntime+1, func(ctx *RunContext) error {
		var handler slog.Handler
		if len(ctx.logHandlers) == 1 {
			handler = ctx.logHandlers[0]
		} else {
			handler = log.NewFanoutHandler(ctx.logHandlers...)
		}
		return sink.Start(sink.NewHandlerSink(handler), nil)
	})
	ExitAt(exitSeqRuntime, func(*RunContext) error {
		_ = sink.Stop(context.Background())
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		return nil
	})
}
