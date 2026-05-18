package gizrun

import (
	"flag"
	"log/slog"
	"net/http"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/cmdhandler"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/configfile"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/log/volclog"
)

type RunContext struct {
	cmdHandler     *cmdhandler.CmdHandler
	httpHandler    http.Handler
	logHandlers    []slog.Handler
	serveMux       *http.ServeMux
	volcLogHandler *volclog.Handler
	configParser   *configfile.YamlParser

	config configfile.ConfigFile

	// flags
	configPath  string
	enablePprof bool
	logLevel    slog.LevelVar
	volcLog     volcLogConfig
}

func (c *RunContext) loadConfig(fs *flag.FlagSet) error {
	config, err := c.configParser.ParseFile(c.configPathValue(fs))
	if err != nil {
		return err
	}
	c.config = config
	return nil
}

func (c *RunContext) Config(name string) (any, bool) {
	if c == nil {
		return nil, false
	}
	return c.config.Config(name)
}

func (c *RunContext) configPathValue(fs *flag.FlagSet) string {
	if c == nil {
		return ""
	}
	if fs == nil {
		return c.configPath
	}
	flag := fs.Lookup("config")
	if flag == nil {
		return c.configPath
	}
	return flag.Value.String()
}
