package gizrun

import (
	"flag"
	"log/slog"
	"strings"
)

func init() {
	InitAt(initSeqFlags, func() error {
		runCtx.logLevel.Set(slog.LevelInfo)
		flag.Var(logLevelFlag{target: &runCtx.logLevel}, "gizrun-log-level", "set gizrun slog level: debug, info, warn, or error")
		return nil
	})
}

type logLevelFlag struct {
	target *slog.LevelVar
}

func (v logLevelFlag) String() string {
	if v.target == nil {
		return slog.LevelInfo.String()
	}
	return v.target.Level().String()
}

func (v logLevelFlag) Set(value string) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(value))); err != nil {
		return err
	}
	v.target.Set(level)
	return nil
}
