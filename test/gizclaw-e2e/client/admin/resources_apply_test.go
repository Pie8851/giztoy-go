//go:build gizclaw_e2e

package admin_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
)

func TestAdminAPIApplyResource(t *testing.T) {
	env := newAdminAPIHarness(t)

	name := mutationName("apply-workflow")
	_, _ = env.api.DeleteWorkflowWithResponse(env.ctx, name)
	var resource apitypes.Resource
	if err := resource.FromWorkflowResource(apitypes.WorkflowResource{
		ApiVersion: apitypes.ResourceAPIVersionGizclawAdminv1alpha1,
		Kind:       apitypes.WorkflowResourceKindWorkflow,
		Metadata:   apitypes.ResourceMetadata{Name: name},
		Spec: apitypes.WorkflowSpec{
			Driver:    apitypes.WorkflowDriverFlowcraft,
			Flowcraft: &apitypes.FlowcraftWorkflowSpec{},
		},
	}); err != nil {
		t.Fatalf("build workflow resource: %v", err)
	}
	resp, err := env.api.ApplyResourceWithResponse(env.ctx, resource)
	if err != nil {
		t.Fatalf("apply resource: %v", err)
	}
	requireStatusOK(t, resp, resp.Body)
	if resp.JSON200 == nil || resp.JSON200.Name != name || resp.JSON200.Kind != apitypes.ResourceKindWorkflow {
		t.Fatalf("apply resource = %#v", resp.JSON200)
	}
	deleted, err := env.api.DeleteWorkflowWithResponse(env.ctx, name)
	if err != nil {
		t.Fatalf("delete applied workflow: %v", err)
	}
	requireStatusOK(t, deleted, deleted.Body)
}
