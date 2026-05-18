package gizrun

import (
	"flag"
	"net/http"
	"net/http/pprof"
)

func init() {
	InitAt(initSeqFlags, func() error {
		flag.BoolVar(&runCtx.enablePprof, "pprof", false, "enable pprof handlers")
		return nil
	})
	PostInitAt(postInitSeqRuntime, func(ctx *RunContext) error {
		if ctx.enablePprof {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			ctx.serveMux.Handle("/debug/pprof/", mux)
		}
		return nil
	})
}
