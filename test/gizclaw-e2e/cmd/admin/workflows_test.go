//go:build gizclaw_e2e

package admin_test

import (
	"strings"
	"testing"

	clitest "github.com/GizClaw/gizclaw-go/test/gizclaw-e2e/cmd"
)

func TestAdminWorkflowsUserStory(t *testing.T) {
	h := clitest.NewSetupHarness(t, "507-admin-workflows")
	h.CreateContext("admin-a").MustSucceed(t)
	h.RegisterContext("admin-a", "--sn", "admin-sn").MustSucceed(t)

	list := h.RunCLI("admin", "workflows", "list", "--context", "admin-a")
	list.MustSucceed(t)
	if !strings.Contains(list.Stdout, `"name":"ui-seed-workflow"`) {
		t.Fatalf("workflows list missing ui-seed-workflow:\n%s", list.Stdout)
	}

	get := h.RunCLI("admin", "workflows", "get", "ui-seed-workflow", "--context", "admin-a")
	get.MustSucceed(t)
	if !strings.Contains(get.Stdout, `"driver":"flowcraft"`) {
		t.Fatalf("workflows get missing driver:\n%s", get.Stdout)
	}
}
