package gizclaw

import (
	"context"
	"net"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestRPCPeerPingServerHandle(t *testing.T) {
	server := &rpcServer{}
	client := &rpcClient{}
	ping := callRPCPair(t, server, func(conn net.Conn) (*rpcapi.PingResponse, error) {
		return client.Ping(context.Background(), conn, "ping")
	})
	if ping.ServerTime <= 0 {
		t.Fatalf("Ping() = %+v", ping)
	}
}

func TestRPCPeerPingClientHandle(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- (&rpcClient{}).Handle(clientSide)
	}()

	ping, err := (&rpcServer{}).Ping(context.Background(), serverSide, "server-ping")
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if ping.ServerTime <= 0 {
		t.Fatalf("Ping() = %+v", ping)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}
