//go:build gizclaw_e2e

package workspace

import "testing"

func TestPushToTalkInterrupt(t *testing.T) {
	runLiveWorkspaceCase(t, workspaceCasePushToTalkInterrupt, allWorkspaceConfigPaths(t))
}
