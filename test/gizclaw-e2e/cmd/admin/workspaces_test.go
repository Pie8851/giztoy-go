//go:build gizclaw_e2e

package admin_test

import (
	"strings"
	"testing"

	clitest "github.com/GizClaw/gizclaw-go/test/gizclaw-e2e/cmd"
)

func TestAdminWorkspacesUserStory(t *testing.T) {
	h := clitest.NewSetupHarness(t, "508-admin-workspaces")
	h.CreateContext("admin-a").MustSucceed(t)
	h.RegisterContext("admin-a", "--sn", "admin-sn").MustSucceed(t)

	list := h.RunCLI("admin", "workspaces", "list", "--context", "admin-a")
	list.MustSucceed(t)
	if !strings.Contains(list.Stdout, `"name":"ui-seed-workspace"`) {
		t.Fatalf("workspaces list missing created item:\n%s", list.Stdout)
	}

	get := h.RunCLI("admin", "workspaces", "get", "ui-seed-workspace", "--context", "admin-a")
	get.MustSucceed(t)
	if !strings.Contains(get.Stdout, `"workflow_name":"ui-seed-workflow"`) {
		t.Fatalf("workspaces get missing workflow name:\n%s", get.Stdout)
	}
}
