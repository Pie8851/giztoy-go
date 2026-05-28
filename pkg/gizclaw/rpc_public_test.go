package gizclaw

import (
	"context"
	"net"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
)

func TestRPCPublicServerInfo(t *testing.T) {
	publicKey := giznet.PublicKey{1, 2, 3}
	serverPublicKey := giznet.PublicKey{9, 8, 7}
	serverInfo := &fakeRPCServerInfoService{
		t:             t,
		wantPublicKey: publicKey,
		info: apitypes.ServerInfo{
			PublicKey:   serverPublicKey.String(),
			ServerTime:  123,
			BuildCommit: "test",
		},
	}
	server := &rpcServer{serverInfo: serverInfo, callerPublicKey: publicKey}
	client := &rpcClient{}

	info := callRPCPair(t, server, func(conn net.Conn) (*rpcapi.ServerGetInfoResponse, error) {
		return client.GetServerInfo(context.Background(), conn, "server-info")
	})
	if info.PublicKey != serverPublicKey.String() || info.ServerTime != 123 || info.BuildCommit != "test" {
		t.Fatalf("GetServerInfo() = %+v", info)
	}
}
