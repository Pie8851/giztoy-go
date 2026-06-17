package agenthost

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/store/objectstore"
)

func TestObjectWorkspaceStorePrepareWorkspaceCreatesLocalDir(t *testing.T) {
	root := t.TempDir()
	store := NewObjectWorkspaceStore(objectstore.Dir(root))

	rt, err := store.PrepareWorkspace(context.Background(), "demo ws")
	if err != nil {
		t.Fatalf("PrepareWorkspace() error = %v", err)
	}
	if rt.ObjectPrefix != "workspaces/demo%20ws" {
		t.Fatalf("ObjectPrefix = %q, want escaped workspace prefix", rt.ObjectPrefix)
	}
	wantDir := filepath.Join(root, "workspaces", "demo%20ws")
	if rt.LocalDir != wantDir {
		t.Fatalf("LocalDir = %q, want %q", rt.LocalDir, wantDir)
	}
	if info, err := os.Stat(wantDir); err != nil || !info.IsDir() {
		t.Fatalf("workspace dir not created: info=%v err=%v", info, err)
	}
}

func TestObjectWorkspaceStorePrepareWorkspaceValidation(t *testing.T) {
	if _, err := (ObjectWorkspaceStore{}).PrepareWorkspace(context.Background(), "demo"); err == nil || !strings.Contains(err.Error(), "workspace store") {
		t.Fatalf("PrepareWorkspace(nil store) error = %v", err)
	}

	store := NewObjectWorkspaceStore(objectstore.Dir(t.TempDir()))
	if _, err := store.PrepareWorkspace(context.Background(), " "); err == nil || !strings.Contains(err.Error(), "workspace name") {
		t.Fatalf("PrepareWorkspace(empty workspace) error = %v", err)
	}
}
