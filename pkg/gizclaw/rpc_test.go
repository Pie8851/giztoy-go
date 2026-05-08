package gizclaw

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpc"
)

func TestRPCClientPingSingleRequestResponse(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	client := newRPCClient(clientSide)
	defer client.Close()

	reqCh := make(chan *rpc.RPCRequest, 1)
	serverErrCh := make(chan error, 1)

	go func() {
		req, err := rpc.ReadRequest(serverSide)
		if err != nil {
			serverErrCh <- err
			return
		}
		reqCh <- req

		resp := &rpc.RPCResponse{
			V:      1,
			Id:     req.Id,
			Result: &rpc.PingResponse{ServerTime: rpcServerTimeForID(req.Id)},
		}
		if err := rpc.WriteResponse(serverSide, resp); err != nil {
			serverErrCh <- err
			return
		}

		serverErrCh <- nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ping, err := client.Ping(ctx, "req-1")
	if err != nil {
		t.Fatalf("Ping(req-1) error: %v", err)
	}
	if ping.ServerTime != rpcServerTimeForID("req-1") {
		t.Fatalf("Ping(req-1) server_time = %d", ping.ServerTime)
	}

	if req := <-reqCh; req.Id == "" {
		t.Fatal("request missing id")
	} else {
		assertRPCPingRequestHasTimestamp(t, req)
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("server goroutine error: %v", err)
	}
}

func TestRPCClientCallContextTimeout(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	client := newRPCClient(clientSide)
	defer client.Close()

	readDone := make(chan *rpc.RPCRequest, 1)
	go func() {
		req, _ := rpc.ReadRequest(serverSide)
		readDone <- req
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.call(ctx, &rpc.RPCRequest{
		V:      1,
		Id:     "timeout",
		Method: rpc.MethodPing,
		Params: &rpc.PingRequest{ClientSendTime: time.Now().UnixMilli()},
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call timeout err = %v, want %v", err, context.DeadlineExceeded)
	}

	if req := <-readDone; req == nil || req.Id != "timeout" {
		t.Fatalf("server received request = %+v", req)
	} else {
		assertRPCPingRequestHasTimestamp(t, req)
	}
}

func TestRPCClientCallContextCancel(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	client := newRPCClient(clientSide)
	defer client.Close()

	readDone := make(chan *rpc.RPCRequest, 1)
	go func() {
		req, _ := rpc.ReadRequest(serverSide)
		readDone <- req
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if req := <-readDone; req != nil {
			cancel()
		}
	}()

	_, err := client.call(ctx, &rpc.RPCRequest{
		V:      1,
		Id:     "cancel",
		Method: rpc.MethodPing,
		Params: &rpc.PingRequest{ClientSendTime: time.Now().UnixMilli()},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Call cancel err = %v, want %v", err, context.Canceled)
	}
}

func TestRPCClientCallValidatesRequestAndClosedState(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()

	client := newRPCClient(clientSide)
	defer client.Close()

	if _, err := client.call(context.Background(), nil); err == nil || err.Error() != "rpc: nil request" {
		t.Fatalf("call(nil) err = %v", err)
	}
	if _, err := client.call(context.Background(), &rpc.RPCRequest{Method: rpc.MethodPing}); err == nil || err.Error() != "rpc: request id required" {
		t.Fatalf("call(empty id) err = %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := client.call(context.Background(), &rpc.RPCRequest{Id: "closed", Method: rpc.MethodPing}); !errors.Is(err, errRPCClientClosed) {
		t.Fatalf("call(closed) err = %v, want %v", err, errRPCClientClosed)
	}
}

func TestRPCClientPingErrorPaths(t *testing.T) {
	t.Run("error response", func(t *testing.T) {
		serverSide, clientSide := net.Pipe()
		defer serverSide.Close()
		defer clientSide.Close()

		client := newRPCClient(clientSide)
		defer client.Close()

		go func() {
			req, _ := rpc.ReadRequest(serverSide)
			_ = rpc.WriteResponse(serverSide, rpc.Error{RequestID: req.Id, Code: -1, Message: "boom"}.RPCResponse())
		}()

		_, err := client.Ping(context.Background(), "ping-error")
		if err == nil || err.Error() != "rpc: boom" {
			t.Fatalf("Ping(error response) err = %v", err)
		}
		var rpcErr rpc.Error
		if !errors.As(err, &rpcErr) {
			t.Fatalf("Ping(error response) err = %T, want rpc.Error", err)
		}
		if rpcErr.RequestID != "ping-error" || rpcErr.Code != -1 || rpcErr.Message != "boom" {
			t.Fatalf("Ping(error response) rpc error = %+v", rpcErr)
		}
	})

	t.Run("missing result", func(t *testing.T) {
		serverSide, clientSide := net.Pipe()
		defer serverSide.Close()
		defer clientSide.Close()

		client := newRPCClient(clientSide)
		defer client.Close()

		go func() {
			req, _ := rpc.ReadRequest(serverSide)
			_ = rpc.WriteResponse(serverSide, &rpc.RPCResponse{V: 1, Id: req.Id})
		}()

		_, err := client.Ping(context.Background(), "ping-missing")
		if err == nil || err.Error() != "rpc: missing ping result" {
			t.Fatalf("Ping(missing result) err = %v", err)
		}
	})
}

func rpcServerTimeForID(id string) int64 {
	if id == "req-1" {
		return 1
	}
	return 2
}

func assertRPCPingRequestHasTimestamp(t *testing.T, req *rpc.RPCRequest) {
	t.Helper()
	if req.Params == nil {
		t.Fatal("ping request params missing")
	}
	if req.Params.ClientSendTime <= 0 {
		t.Fatalf("ping request client_send_time = %d", req.Params.ClientSendTime)
	}
}
