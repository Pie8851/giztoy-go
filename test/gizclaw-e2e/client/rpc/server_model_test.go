//go:build gizclaw_e2e

package rpc_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerModelRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	modelList, err := env.peer.ListModels(env.ctx, "model.list.seeded", rpcapi.ModelListRequest{})
	if err != nil {
		t.Fatalf("model.list seeded: %v", err)
	}
	if !hasModel(modelList.Items, "seed-model") {
		t.Fatalf("model.list missing seed-model: %#v", modelList.Items)
	}
	seedModel, err := env.peer.GetModel(env.ctx, "model.get.seeded", rpcapi.ModelGetRequest{Id: "seed-model"})
	if err != nil {
		t.Fatalf("model.get seeded: %v", err)
	}
	if seedModel.Id != "seed-model" {
		t.Fatalf("model.get seeded id = %q", seedModel.Id)
	}
	model, err := env.peer.CreateModel(env.ctx, "model.create", rpcModel("peer-model", "peer-provider"))
	if err != nil {
		t.Fatalf("model.create: %v", err)
	}
	if model.Id != "peer-model" {
		t.Fatalf("model.create id = %q", model.Id)
	}
	modelName := "peer model updated"
	model, err = env.peer.PutModel(env.ctx, "model.put", rpcapi.ModelPutRequest{
		Id: "peer-model",
		Body: func() rpcapi.Model {
			body := rpcModel("peer-model", "peer-provider")
			body.Name = &modelName
			return body
		}(),
	})
	if err != nil {
		t.Fatalf("model.put: %v", err)
	}
	if model.Name == nil || *model.Name != modelName {
		t.Fatalf("model.put name = %#v", model.Name)
	}
	model, err = env.peer.GetModel(env.ctx, "model.get.updated", rpcapi.ModelGetRequest{Id: "peer-model"})
	if err != nil {
		t.Fatalf("model.get updated: %v", err)
	}
	if model.Name == nil || *model.Name != modelName {
		t.Fatalf("model.get updated name = %#v", model.Name)
	}
	assertModelPagination(t, env.ctx, env.peer, "seed-model", "peer-model")
	if _, err := env.peer.DeleteModel(env.ctx, "model.delete", rpcapi.ModelDeleteRequest{Id: "peer-model"}); err != nil {
		t.Fatalf("model.delete: %v", err)
	}
}
