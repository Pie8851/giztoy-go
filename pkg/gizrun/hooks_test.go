package gizrun

import (
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/cmdhandler"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/configfile"
)

func TestRunParsesCommandFlagsOnly(t *testing.T) {
	resetInitHooksForTest(t)
	resetDefaultCmdHandlerForTest(t)
	resetFlagSetForTest(t)
	configFile := filepath.Join(t.TempDir(), "app.yaml")
	if err := os.WriteFile(configFile, nil, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	previousArgs := os.Args
	os.Args = []string{"gizrun", "-config", configFile, "chat", "-model=gpt"}
	t.Cleanup(func() {
		os.Args = previousArgs
	})

	var value string
	var gotArgs []string
	var gotFlags []string
	InitAt(0, func() error {
		flag.StringVar(&value, "config", "", "config path")
		return nil
	})
	if err := HandleCmd("chat", CmdHandleFunc(func(_ context.Context, args []string, flags []string) error {
		gotArgs = append([]string(nil), args...)
		gotFlags = append([]string(nil), flags...)
		return nil
	})); err != nil {
		t.Fatalf("HandleCmd failed: %v", err)
	}

	if err := Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if value != configFile {
		t.Fatalf("config flag = %q, want %q", value, configFile)
	}
	if len(gotArgs) != 0 {
		t.Fatalf("args = %#v, want empty", gotArgs)
	}
	if len(gotFlags) != 1 || gotFlags[0] != "-model=gpt" {
		t.Fatalf("flags = %#v, want model flag", gotFlags)
	}
}

func TestRuntimeHooksInitializeHTTPAndPprof(t *testing.T) {
	initHookValues := append([]initHook(nil), initHooks.hooks...)
	postInitHookValues := append([]postInitHook(nil), postInitHooks.hooks...)
	exitHookValues := append([]exitHook(nil), exitHooks.hooks...)
	resetRuntimeForTest(t)
	resetFlagSetForTest(t)

	runInitHooks(initHookValues)
	if err := flag.CommandLine.Set("pprof", "true"); err != nil {
		t.Fatalf("set pprof flag: %v", err)
	}
	if err := runCtx.loadConfig(flag.CommandLine); err != nil {
		t.Fatalf("load config: %v", err)
	}
	runPostInitHooks(&runCtx, postInitHookValues)
	t.Cleanup(func() {
		runExitHooks(&runCtx, exitHookValues)
	})

	metricsRec := httptest.NewRecorder()
	runCtx.httpHandler.ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metricsRec.Code != http.StatusOK {
		t.Fatalf("/metrics status = %d, want 200", metricsRec.Code)
	}
	pprofRec := httptest.NewRecorder()
	runCtx.httpHandler.ServeHTTP(pprofRec, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if pprofRec.Code != http.StatusOK {
		t.Fatalf("/debug/pprof/ status = %d, want 200", pprofRec.Code)
	}
}

func TestPostInitAt(t *testing.T) {
	resetInitHooksForTest(t)
	postInitHooks.next = 0
	postInitHooks.hooks = nil

	var got []int
	PostInitAt(0x20, func(*RunContext) error { got = append(got, 2); return nil })
	PostInitAt(0x10, func(*RunContext) error { got = append(got, 1); return nil })
	PostInitAt(0x20, func(*RunContext) error { got = append(got, 3); return nil })
	PostInitAt(0x30, nil)

	runInitHooks(initHooks.hooks)
	runPostInitHooks(&runCtx, postInitHooks.hooks)

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("post-init hook count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("post-init hook order = %#v, want %#v", got, want)
		}
	}
}

func TestInitAt(t *testing.T) {
	resetInitHooksForTest(t)

	var got []int
	InitAt(0x20, func() error { got = append(got, 2); return nil })
	InitAt(0x10, func() error { got = append(got, 1); return nil })
	InitAt(0x20, func() error { got = append(got, 3); return nil })
	InitAt(0x30, nil)

	runInitHooks(initHooks.hooks)

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("init hook count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("init hook order = %#v, want %#v", got, want)
		}
	}
}

func TestExitAt(t *testing.T) {
	resetInitHooksForTest(t)
	exitHooks.next = 0
	exitHooks.hooks = nil

	var got []int
	ExitAt(0x20, func(*RunContext) error { got = append(got, 2); return nil })
	ExitAt(0x10, func(*RunContext) error { got = append(got, 1); return nil })
	ExitAt(0x20, func(*RunContext) error { got = append(got, 3); return nil })
	ExitAt(0x30, nil)

	runExitHooks(&runCtx, exitHooks.hooks)

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("exit hook count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("exit hook order = %#v, want %#v", got, want)
		}
	}
}

func resetInitHooksForTest(t *testing.T) {
	t.Helper()
	previousInit := initHooks
	previousPostInit := postInitHooks
	previousExit := exitHooks
	t.Cleanup(func() {
		initHooks = previousInit
		postInitHooks = previousPostInit
		exitHooks = previousExit
	})
	initHooks.next = 1
	initHooks.hooks = []initHook{{
		seq: 0,
		fn: func() error {
			return nil
		},
	}}
	postInitHooks.next = 1
	postInitHooks.hooks = []postInitHook{{
		seq: 0,
		fn:  func(*RunContext) error { return nil },
	}}
	exitHooks.next = 1
	exitHooks.hooks = []exitHook{{
		seq: 0,
		fn:  func(*RunContext) error { return nil },
	}}
}

func resetRuntimeForTest(t *testing.T) {
	t.Helper()
	previous := runCtx
	runCtx = RunContext{
		cmdHandler:   cmdhandler.New(),
		configParser: configfile.NewYamlParser(),
		serveMux:     http.NewServeMux(),
	}
	t.Cleanup(func() {
		runCtx = previous
	})
}

func resetFlagSetForTest(t *testing.T) {
	t.Helper()
	previous := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("gizrun-test", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	t.Cleanup(func() {
		flag.CommandLine = previous
	})
}
