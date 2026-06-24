//go:build gizclaw_e2e

package workspace

import "testing"

func TestRealtimeAutoSplitHistory(t *testing.T) {
	runLiveWorkspaceCase(t, workspaceCaseRealtimeAutoSplit, allWorkspaceConfigPaths(t))
}
