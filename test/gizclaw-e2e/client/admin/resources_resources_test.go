//go:build gizclaw_e2e

package admin_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
)

func TestAdminAPIResourcesGet(t *testing.T) {
	env := newAdminAPIHarness(t)

	get, err := env.api.GetResourceWithResponse(env.ctx, apitypes.ResourceKindWorkflow, "e2e-rpc-workflow")
	if err != nil {
		t.Fatalf("get workflow resource: %v", err)
	}
	requireStatusOK(t, get, get.Body)
	if get.JSON200 == nil {
		t.Fatalf("get resource = %#v", get.JSON200)
	}
	workflow, err := get.JSON200.AsWorkflowResource()
	if err != nil {
		t.Fatalf("decode workflow resource union: %v", err)
	}
	if workflow.Metadata.Name != "e2e-rpc-workflow" {
		t.Fatalf("workflow resource name = %q", workflow.Metadata.Name)
	}
}
