//go:build gizclaw_e2e

package rpc_test

import (
	"strings"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerWorkspaceRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	if _, err := env.peer.CreateWorkflow(env.ctx, "workspace.workflow.create", rpcWorkflow("peer-flow", "workspace test flow")); err != nil {
		t.Fatalf("create workflow for workspace test: %v", err)
	}
	workspaceList, err := env.peer.ListWorkspaces(env.ctx, "workspace.list.seeded", rpcapi.WorkspaceListRequest{})
	if err != nil {
		t.Fatalf("workspace.list seeded: %v", err)
	}
	if !hasWorkspace(workspaceList.Items, "seed-workspace") {
		t.Fatalf("workspace.list missing seed-workspace: %#v", workspaceList.Items)
	}
	assertWorkspacePrefixList(t, env.ctx, env.peer)
	seedWorkspace, err := env.peer.GetWorkspace(env.ctx, "workspace.get.seeded", rpcapi.WorkspaceGetRequest{Name: "seed-workspace"})
	if err != nil {
		t.Fatalf("workspace.get seeded: %v", err)
	}
	if seedWorkspace.Name != "seed-workspace" || seedWorkspace.WorkflowName != "seed-flow" {
		t.Fatalf("workspace.get seeded = %#v", seedWorkspace)
	}

	createInput := rpcapi.WorkspaceInputModePushToTalk
	var createParams rpcapi.WorkspaceParameters
	if err := createParams.FromFlowcraftWorkspaceParameters(rpcapi.FlowcraftWorkspaceParameters{Input: &createInput}); err != nil {
		t.Fatalf("FromFlowcraftWorkspaceParameters(create) error = %v", err)
	}
	workspace, err := env.peer.CreateWorkspace(env.ctx, "workspace.create", rpcapi.WorkspaceCreateRequest{
		Name:         "peer-workspace",
		WorkflowName: "peer-flow",
		Parameters:   &createParams,
	})
	if err != nil {
		t.Fatalf("workspace.create: %v", err)
	}
	if workspace.Name != "peer-workspace" || workspace.WorkflowName != "peer-flow" {
		t.Fatalf("workspace.create = %#v", workspace)
	}

	updateInput := rpcapi.WorkspaceInputModeRealtime
	var updateParams rpcapi.WorkspaceParameters
	if err := updateParams.FromFlowcraftWorkspaceParameters(rpcapi.FlowcraftWorkspaceParameters{Input: &updateInput}); err != nil {
		t.Fatalf("FromFlowcraftWorkspaceParameters(update) error = %v", err)
	}
	workspace, err = env.peer.PutWorkspace(env.ctx, "workspace.put", rpcapi.WorkspacePutRequest{
		Name: "peer-workspace",
		Body: rpcapi.Workspace{
			Name:         "peer-workspace",
			WorkflowName: "peer-flow",
			Parameters:   &updateParams,
		},
	})
	if err != nil {
		t.Fatalf("workspace.put: %v", err)
	}
	typed, err := workspace.Parameters.AsFlowcraftWorkspaceParameters()
	if err != nil {
		t.Fatalf("workspace.put parameters decode: %v", err)
	}
	if typed.Input == nil || *typed.Input != rpcapi.WorkspaceInputModeRealtime {
		t.Fatalf("workspace.put input = %#v, want realtime", typed.Input)
	}
	workspace, err = env.peer.GetWorkspace(env.ctx, "workspace.get.updated", rpcapi.WorkspaceGetRequest{Name: "peer-workspace"})
	if err != nil {
		t.Fatalf("workspace.get updated: %v", err)
	}
	typed, err = workspace.Parameters.AsFlowcraftWorkspaceParameters()
	if err != nil {
		t.Fatalf("workspace.get updated parameters decode: %v", err)
	}
	if typed.Input == nil || *typed.Input != rpcapi.WorkspaceInputModeRealtime {
		t.Fatalf("workspace.get updated input = %#v, want realtime", typed.Input)
	}
	assertWorkspacePagination(t, env.ctx, env.peer, "seed-workspace", "peer-workspace")
	if _, err := env.peer.DeleteWorkspace(env.ctx, "workspace.delete", rpcapi.WorkspaceDeleteRequest{Name: "peer-workspace"}); err != nil {
		t.Fatalf("workspace.delete: %v", err)
	}
	if _, err := env.peer.DeleteWorkflow(env.ctx, "workspace.workflow.delete", rpcapi.WorkflowDeleteRequest{Name: "peer-flow"}); err != nil {
		t.Fatalf("delete workflow for workspace test: %v", err)
	}
}

func TestServerResourceACLRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	denied := env.h.ConnectClientFromContext("peer-denied")
	defer denied.Close()
	if _, err := denied.GetWorkflow(env.ctx, "workflow.get.denied", rpcapi.WorkflowGetRequest{Name: "seed-flow"}); err == nil || !strings.Contains(err.Error(), "acl: denied") {
		t.Fatalf("denied peer workflow.get error = %v", err)
	}
	if _, err := denied.GetWorkspace(env.ctx, "workspace.get.denied", rpcapi.WorkspaceGetRequest{Name: "seed-workspace"}); err == nil || !strings.Contains(err.Error(), "acl: denied") {
		t.Fatalf("denied peer workspace.get error = %v", err)
	}
	if _, err := denied.GetModel(env.ctx, "model.get.denied", rpcapi.ModelGetRequest{Id: "seed-model"}); err == nil || !strings.Contains(err.Error(), "acl: denied") {
		t.Fatalf("denied peer model.get error = %v", err)
	}
	if _, err := denied.GetCredential(env.ctx, "credential.get.denied", rpcapi.CredentialGetRequest{Name: "seed-credential"}); err == nil || !strings.Contains(err.Error(), "acl: denied") {
		t.Fatalf("denied peer credential.get error = %v", err)
	}
	assertDeniedListsAreEmpty(t, env.ctx, denied)
}
