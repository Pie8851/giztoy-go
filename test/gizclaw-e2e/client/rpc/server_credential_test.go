//go:build gizclaw_e2e

package rpc_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerCredentialRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	credentialList, err := env.peer.ListCredentials(env.ctx, "credential.list.seeded", rpcapi.CredentialListRequest{})
	if err != nil {
		t.Fatalf("credential.list seeded: %v", err)
	}
	if !hasCredential(credentialList.Items, "seed-credential") {
		t.Fatalf("credential.list missing seed-credential: %#v", credentialList.Items)
	}
	seedCredential, err := env.peer.GetCredential(env.ctx, "credential.get.seeded", rpcapi.CredentialGetRequest{Name: "seed-credential"})
	if err != nil {
		t.Fatalf("credential.get seeded: %v", err)
	}
	if seedCredential.Name != "seed-credential" {
		t.Fatalf("credential.get seeded name = %q", seedCredential.Name)
	}
	credential, err := env.peer.CreateCredential(env.ctx, "credential.create", rpcCredential("peer-credential", "sk-created"))
	if err != nil {
		t.Fatalf("credential.create: %v", err)
	}
	if credential.Name != "peer-credential" {
		t.Fatalf("credential.create name = %q", credential.Name)
	}
	credential, err = env.peer.PutCredential(env.ctx, "credential.put", rpcapi.CredentialPutRequest{
		Name: "peer-credential",
		Body: rpcCredential("peer-credential", "sk-updated"),
	})
	if err != nil {
		t.Fatalf("credential.put: %v", err)
	}
	if testRPCCredentialBodyString(credential.Body, "api_key") != "sk-updated" {
		t.Fatalf("credential.put body = %#v", credential.Body)
	}
	credential, err = env.peer.GetCredential(env.ctx, "credential.get.updated", rpcapi.CredentialGetRequest{Name: "peer-credential"})
	if err != nil {
		t.Fatalf("credential.get updated: %v", err)
	}
	if testRPCCredentialBodyString(credential.Body, "api_key") != "sk-updated" {
		t.Fatalf("credential.get updated body = %#v", credential.Body)
	}
	assertCredentialPagination(t, env.ctx, env.peer, "seed-credential", "peer-credential")
	if _, err := env.peer.DeleteCredential(env.ctx, "credential.delete", rpcapi.CredentialDeleteRequest{Name: "peer-credential"}); err != nil {
		t.Fatalf("credential.delete: %v", err)
	}
}
