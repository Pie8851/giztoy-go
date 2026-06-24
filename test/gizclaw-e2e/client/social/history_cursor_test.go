//go:build gizclaw_e2e

package social_test

import "testing"

func TestSocialHistoryCursorRPC(t *testing.T) {
	h := newSocialSimulatorHarness(t)
	peerB := h.ContextPublicKey("peer-b")

	requestAB := createAcceptedFriendRequest(t, h, "peer-a", "peer-b", peerB, "123456")
	assertChatWorkspaceHistory(t, h, "peer-a", "peer-b", stringValue(requestAB.WorkspaceName), []string{
		"hello cursor round one",
		"hello cursor round two",
		"hello cursor round three",
	})
}
