//go:build gizclaw_e2e

package workspace

import "testing"

func TestHistoryReplay(t *testing.T) {
	runLiveWorkspaceCase(t, workspaceCaseHistoryReplay, allWorkspaceConfigPaths(t))
}
