package server

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GizClaw/gizclaw-go/cmd/internal/storage"
	"github.com/GizClaw/gizclaw-go/cmd/internal/stores"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
	"github.com/GizClaw/gizclaw-go/pkg/store/kv"
)

func TestNewMigratorRunsACLMigration(t *testing.T) {
	migrator, err := NewMigrator(validLayeredConfig(t.TempDir()))
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}
	defer migrator.Close()

	for range 2 {
		if err := migrator.Migrate(context.Background()); err != nil {
			t.Fatalf("Migrate() error = %v", err)
		}
	}
	if _, err := migrator.ACL.DB.ExecContext(context.Background(), `INSERT INTO acl_views (name, created_at, updated_at) VALUES ('default', 'now', 'now')`); err != nil {
		t.Fatalf("insert acl view after migration: %v", err)
	}
}

func TestCmdMigratorCloseHandlesNilState(t *testing.T) {
	var nilMigrator *CmdMigrator
	if err := nilMigrator.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
	if err := (&CmdMigrator{}).Close(); err != nil {
		t.Fatalf("empty Close() error = %v", err)
	}
}

func TestMigrateWorkspaceRunsACLMigrationFromWorkspaceConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspaceConfigFile), []byte(`
listen: "127.0.0.1:0"
storage:
  acl-db:
    kind: sql
    sqlite:
      dir: data/acl.sqlite
stores:
  acl:
    kind: sql
    storage: acl-db
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := MigrateWorkspace(context.Background(), root); err != nil {
		t.Fatalf("MigrateWorkspace() error = %v", err)
	}

	migrator, err := NewMigrator(Config{
		Storage: map[string]storage.Config{
			"acl-db": {Kind: storage.KindSQL, SQLite: &storage.SQLConfig{Dir: filepath.Join(root, "data", "acl.sqlite")}},
		},
		Stores: map[string]stores.Config{
			"acl": {Kind: stores.KindSQL, Storage: "acl-db"},
		},
	})
	if err != nil {
		t.Fatalf("NewMigrator() after workspace migration error = %v", err)
	}
	defer migrator.Close()
	if _, err := migrator.ACL.DB.ExecContext(context.Background(), `INSERT INTO acl_views (name, created_at, updated_at) VALUES ('workspace', 'now', 'now')`); err != nil {
		t.Fatalf("insert acl view after workspace migration: %v", err)
	}
}

func TestMigrateWorkspaceMigratesLegacyPeerRole(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspaceConfigFile), []byte(`
listen: "127.0.0.1:0"
storage:
  peer-db:
    kind: keyvalue
    badger:
      dir: data/peer.badger
  acl-db:
    kind: sql
    sqlite:
      dir: data/acl.sqlite
stores:
  peers:
    kind: keyvalue
    storage: peer-db
  acl:
    kind: sql
    storage: acl-db
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := prepareWorkspaceMigrationConfig(root)
	if err != nil {
		t.Fatalf("prepareWorkspaceMigrationConfig() error = %v", err)
	}
	ss, err := newStoreRegistry(cfg)
	if err != nil {
		t.Fatalf("newStoreRegistry() error = %v", err)
	}
	peerStore, err := ss.KV(defaultPeersStore)
	if err != nil {
		t.Fatalf("KV(peers) error = %v", err)
	}
	publicKey := giznet.PublicKey{1}
	legacy := apitypes.Peer{
		PublicKey:     publicKey.String(),
		Role:          apitypes.PeerRole("gear"),
		Status:        apitypes.PeerRegistrationStatusActive,
		Device:        apitypes.DeviceInfo{},
		Configuration: apitypes.Configuration{},
		CreatedAt:     time.Unix(1, 0).UTC(),
		UpdatedAt:     time.Unix(1, 0).UTC(),
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy peer: %v", err)
	}
	if err := peerStore.BatchSet(context.Background(), []kv.Entry{
		{Key: kv.Key{"by-pubkey", legacy.PublicKey}, Value: data},
		{Key: kv.Key{"by-role", "gear", legacy.PublicKey}, Value: []byte{1}},
	}); err != nil {
		t.Fatalf("seed legacy peer: %v", err)
	}
	if err := ss.Close(); err != nil {
		t.Fatalf("close seed stores: %v", err)
	}

	if err := MigrateWorkspace(context.Background(), root); err != nil {
		t.Fatalf("MigrateWorkspace() error = %v", err)
	}

	migrator, err := NewMigrator(cfg)
	if err != nil {
		t.Fatalf("NewMigrator() after migration error = %v", err)
	}
	defer migrator.Close()
	migrated, err := migrator.Peers.LoadPeer(context.Background(), publicKey)
	if err != nil {
		t.Fatalf("LoadPeer() error = %v", err)
	}
	if migrated.Role != apitypes.PeerRoleClient {
		t.Fatalf("Role = %q, want %q", migrated.Role, apitypes.PeerRoleClient)
	}
	if _, err := migrator.Peers.Store.Get(context.Background(), kv.Key{"by-role", "gear", legacy.PublicKey}); !errors.Is(err, kv.ErrNotFound) {
		t.Fatalf("legacy role index err = %v, want %v", err, kv.ErrNotFound)
	}
	if _, err := migrator.Peers.Store.Get(context.Background(), kv.Key{"by-role", "client", legacy.PublicKey}); err != nil {
		t.Fatalf("client role index missing: %v", err)
	}
}

func TestMigrateWorkspaceMigratesLegacyCredentialMethod(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspaceConfigFile), []byte(`
listen: "127.0.0.1:0"
storage:
  credential-db:
    kind: keyvalue
    badger:
      dir: data/credential.badger
  acl-db:
    kind: sql
    sqlite:
      dir: data/acl.sqlite
stores:
  credentials:
    kind: keyvalue
    storage: credential-db
  acl:
    kind: sql
    storage: acl-db
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := prepareWorkspaceMigrationConfig(root)
	if err != nil {
		t.Fatalf("prepareWorkspaceMigrationConfig() error = %v", err)
	}
	ss, err := newStoreRegistry(cfg)
	if err != nil {
		t.Fatalf("newStoreRegistry() error = %v", err)
	}
	credentialStore, err := ss.KV(defaultCredentialsStore)
	if err != nil {
		t.Fatalf("KV(credentials) error = %v", err)
	}
	legacy := []byte(`{
		"name":"legacy-openai",
		"provider":"openai",
		"method":"api_key",
		"body":{"method":"api_key","api_key":"sk-old"},
		"created_at":"2026-01-01T00:00:00Z",
		"updated_at":"2026-01-01T00:00:00Z"
	}`)
	if err := credentialStore.BatchSet(context.Background(), []kv.Entry{
		{Key: kv.Key{"by-name", "legacy-openai"}, Value: legacy},
	}); err != nil {
		t.Fatalf("seed legacy credential: %v", err)
	}
	if err := ss.Close(); err != nil {
		t.Fatalf("close seed stores: %v", err)
	}

	if err := MigrateWorkspace(context.Background(), root); err != nil {
		t.Fatalf("MigrateWorkspace() error = %v", err)
	}

	ss, err = newStoreRegistry(cfg)
	if err != nil {
		t.Fatalf("newStoreRegistry() after migration error = %v", err)
	}
	defer ss.Close()
	credentialStore, err = ss.KV(defaultCredentialsStore)
	if err != nil {
		t.Fatalf("KV(credentials) after migration error = %v", err)
	}
	data, err := credentialStore.Get(context.Background(), kv.Key{"by-name", "legacy-openai"})
	if err != nil {
		t.Fatalf("get migrated credential: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("decode migrated credential: %v", err)
	}
	if _, ok := raw["method"]; ok {
		t.Fatalf("migrated credential still has method: %s", data)
	}
	body, _ := raw["body"].(map[string]any)
	if _, ok := body["method"]; ok {
		t.Fatalf("migrated credential body still has method: %s", data)
	}
	if body["api_key"] != "sk-old" {
		t.Fatalf("api_key = %#v, want sk-old", body["api_key"])
	}
	if _, err := credentialStore.Get(context.Background(), kv.Key{"by-provider", "openai", "legacy-openai"}); err != nil {
		t.Fatalf("provider index missing: %v", err)
	}
}

func TestNewMigratorRequiresACLLogicalStore(t *testing.T) {
	if _, err := NewMigrator(Config{}); err == nil {
		t.Fatal("NewMigrator() error = nil")
	}
}
