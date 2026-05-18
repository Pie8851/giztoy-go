package gizrun

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/cmdhandler"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/configfile"
)

const (
	initSeqFlags       = 0
	postInitSeqRuntime = 0
	exitSeqRuntime     = 0
)

var runCtx = RunContext{
	cmdHandler:   cmdhandler.New(),
	configParser: configfile.NewYamlParser(),
	serveMux:     http.NewServeMux(),
}

type (
	CmdHandler    = cmdhandler.Handler
	CmdHandleFunc = cmdhandler.HandleFunc
	ConfigParser  = configfile.Parser
)

func init() {
	InitAt(initSeqFlags, func() error {
		flag.StringVar(&runCtx.configPath, "config", runCtx.configPath, "config file path")
		return nil
	})
	PostInitAt(postInitSeqRuntime, func(ctx *RunContext) error {
		ctx.httpHandler = wrapHandler(ctx.serveMux)
		return nil
	})
}

func HandleCmd(path string, handler CmdHandler) error {
	return runCtx.cmdHandler.Handle(path, handler)
}

func RegisterConfigParser(name string, parser ConfigParser) {
	runCtx.configParser.Register(name, parser)
}

func Context() *RunContext {
	return &runCtx
}

func Run() error {
	runInitHooks(initHooks.hooks)
	args, flags, err := cmdhandler.Parse(flag.CommandLine, os.Args[1:])
	if err != nil {
		return err
	}
	if err := runCtx.loadConfig(flag.CommandLine); err != nil {
		return err
	}
	handler, ok := runCtx.cmdHandler.Lookup(strings.Join(args, "/"))
	if !ok {
		return errors.New("gizrun: command handler not found")
	}

	runPostInitHooks(&runCtx, postInitHooks.hooks)
	defer runExitHooks(&runCtx, exitHooks.hooks)
	return handler.Handle(context.Background(), nil, flags)
}
