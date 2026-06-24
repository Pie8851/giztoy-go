//go:build gizclaw_e2e

package workspace

import "testing"

func TestRealtimeInterrupt(t *testing.T) {
	runLiveWorkspaceCase(t, workspaceCaseRealtimeInterrupt, allWorkspaceConfigPaths(t))
}
