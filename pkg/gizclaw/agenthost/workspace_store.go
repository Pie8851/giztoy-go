package agenthost

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/GizClaw/gizclaw-go/pkg/store/objectstore"
)

type WorkspaceRuntime struct {
	ObjectPrefix string
	LocalDir     string
}

type WorkspaceStore interface {
	PrepareWorkspace(context.Context, string) (WorkspaceRuntime, error)
}

type ObjectWorkspaceStore struct {
	Objects objectstore.ObjectStore
}

func NewObjectWorkspaceStore(objects objectstore.ObjectStore) ObjectWorkspaceStore {
	return ObjectWorkspaceStore{Objects: objects}
}

func (s ObjectWorkspaceStore) PrepareWorkspace(_ context.Context, workspace string) (WorkspaceRuntime, error) {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return WorkspaceRuntime{}, fmt.Errorf("agenthost: workspace name is required")
	}
	if s.Objects == nil {
		return WorkspaceRuntime{}, fmt.Errorf("agenthost: workspace store is required")
	}
	objectPrefix := workspaceObjectPrefix(workspace)
	rt := WorkspaceRuntime{ObjectPrefix: objectPrefix}
	if provider, ok := s.Objects.(objectstore.LocalDirProvider); ok {
		root, ok := provider.LocalDir()
		if ok && strings.TrimSpace(root) != "" {
			rt.LocalDir = filepath.Join(root, filepath.FromSlash(objectPrefix))
			if err := os.MkdirAll(rt.LocalDir, 0o755); err != nil {
				return WorkspaceRuntime{}, fmt.Errorf("agenthost: create workspace runtime dir: %w", err)
			}
		}
	}
	return rt, nil
}

func workspaceObjectPrefix(workspace string) string {
	return "workspaces/" + url.PathEscape(workspace)
}
