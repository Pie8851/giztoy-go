//go:build gizclaw_e2e

package social_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestSocialFriendGroupMemberRPC(t *testing.T) {
	h := newSocialSimulatorHarness(t)
	peerB := h.ContextPublicKey("peer-b")
	peerC := h.ContextPublicKey("peer-c")

	group := mustCreateFriendGroup(t, h, "peer-a", "family", "voice room")
	memberB := mustAddFriendGroupMember(t, h, "peer-a", stringValue(group.Id), peerB, rpcapi.FriendGroupMemberMutableRoleMember)
	if stringValue(memberB.PeerId) != peerB {
		t.Fatalf("member b peer_id = %q, want %q", stringValue(memberB.PeerId), peerB)
	}
	memberB = mustPutFriendGroupMember(t, h, "peer-a", stringValue(group.Id), peerB, rpcapi.FriendGroupMemberMutableRoleAdmin)
	if memberB.Role == nil || *memberB.Role != rpcapi.FriendGroupMemberRoleAdmin {
		t.Fatalf("member b role = %v, want admin", memberB.Role)
	}
	memberC := mustAddFriendGroupMember(t, h, "peer-b", stringValue(group.Id), peerC, rpcapi.FriendGroupMemberMutableRoleMember)
	if stringValue(memberC.PeerId) != peerC {
		t.Fatalf("member c peer_id = %q, want %q", stringValue(memberC.PeerId), peerC)
	}
	assertFriendGroupMemberPagination(t, h, stringValue(group.Id))

	deletedMember := mustDeleteFriendGroupMember(t, h, "peer-b", stringValue(group.Id), peerC)
	if stringValue(deletedMember.PeerId) != peerC {
		t.Fatalf("friend_group.members.delete peer_id = %q, want %q", stringValue(deletedMember.PeerId), peerC)
	}
	assertWorkspaceHistoryDenied(t, h, "peer-c", stringValue(group.WorkspaceName))
}
