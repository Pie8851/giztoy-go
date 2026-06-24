//go:build gizclaw_e2e

package social_test

import "testing"

func TestSocialFriendRequestRPC(t *testing.T) {
	h := newSocialSimulatorHarness(t)
	peerB := h.ContextPublicKey("peer-b")
	peerC := h.ContextPublicKey("peer-c")

	assertFriendOTPFailureCases(t, h, peerB)
	requestAB := createAcceptedFriendRequest(t, h, "peer-a", "peer-b", peerB, "123456")
	requestAC := createAcceptedFriendRequest(t, h, "peer-a", "peer-c", peerC, "234567")
	if stringValue(requestAB.WorkspaceName) == "" || stringValue(requestAC.WorkspaceName) == "" {
		t.Fatalf("accepted friend workspaces are empty: ab=%#v ac=%#v", requestAB, requestAC)
	}
	assertFriendPagination(t, h, requestAB, requestAC)
	assertRejectedFriendRequest(t, h, peerB)

	deletedFriend := mustDeleteFriend(t, h, "peer-a", stringValue(requestAC.Id))
	if stringValue(deletedFriend.Id) != stringValue(requestAC.Id) {
		t.Fatalf("friend.delete id = %q, want %q", stringValue(deletedFriend.Id), stringValue(requestAC.Id))
	}
	assertWorkspaceHistoryDenied(t, h, "peer-c", stringValue(requestAC.WorkspaceName))
}
