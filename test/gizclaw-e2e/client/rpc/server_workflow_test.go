//go:build gizclaw_e2e

package rpc_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerWorkflowRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	workflowList, err := env.peer.ListWorkflows(env.ctx, "workflow.list.seeded", rpcapi.WorkflowListRequest{})
	if err != nil {
		t.Fatalf("workflow.list seeded: %v", err)
	}
	if !hasWorkflow(workflowList.Items, "seed-flow") {
		t.Fatalf("workflow.list missing seed-flow: %#v", workflowList.Items)
	}
	seedFlow, err := env.peer.GetWorkflow(env.ctx, "workflow.get.seeded", rpcapi.WorkflowGetRequest{Name: "seed-flow"})
	if err != nil {
		t.Fatalf("workflow.get seeded: %v", err)
	}
	if seedFlow.Metadata.Name != "seed-flow" {
		t.Fatalf("workflow.get seeded name = %q", seedFlow.Metadata.Name)
	}

	createdFlow, err := env.peer.CreateWorkflow(env.ctx, "workflow.create", rpcWorkflow("peer-flow", "created by peer rpc"))
	if err != nil {
		t.Fatalf("workflow.create: %v", err)
	}
	if createdFlow.Metadata.Name != "peer-flow" {
		t.Fatalf("workflow.create name = %q", createdFlow.Metadata.Name)
	}
	updatedFlowDoc := rpcWorkflow("peer-flow", "updated by peer rpc")
	updatedFlow, err := env.peer.PutWorkflow(env.ctx, "workflow.put", rpcapi.WorkflowPutRequest{Name: "peer-flow", Body: updatedFlowDoc})
	if err != nil {
		t.Fatalf("workflow.put: %v", err)
	}
	if updatedFlow.Metadata.Description == nil || *updatedFlow.Metadata.Description != "updated by peer rpc" {
		t.Fatalf("workflow.put description = %#v", updatedFlow.Metadata.Description)
	}
	assertWorkflowPagination(t, env.ctx, env.peer, "seed-flow", "peer-flow")
	if _, err := env.peer.DeleteWorkflow(env.ctx, "workflow.delete", rpcapi.WorkflowDeleteRequest{Name: "peer-flow"}); err != nil {
		t.Fatalf("workflow.delete: %v", err)
	}
}
