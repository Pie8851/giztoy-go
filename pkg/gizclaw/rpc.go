package gizclaw

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpc"
)

var errRPCClientClosed = errors.New("rpc: client closed")

type rpcClient struct {
	conn net.Conn

	callMu    sync.Mutex
	closeOnce sync.Once
	closed    atomic.Bool
}

func newRPCClient(conn net.Conn) *rpcClient {
	return &rpcClient{conn: conn}
}

func (c *rpcClient) call(ctx context.Context, req *rpc.RPCRequest) (*rpc.RPCResponse, error) {
	if req == nil {
		return nil, errors.New("rpc: nil request")
	}
	if req.Id == "" {
		return nil, errors.New("rpc: request id required")
	}
	if c.closed.Load() {
		return nil, errRPCClientClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.callMu.Lock()
	defer c.callMu.Unlock()
	if c.closed.Load() {
		return nil, errRPCClientClosed
	}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		if err := c.conn.SetDeadline(deadline); err != nil {
			return nil, err
		}
	}
	stopCancel := make(chan struct{})
	cancelDone := make(chan struct{})
	defer func() {
		close(stopCancel)
		<-cancelDone
		_ = c.conn.SetDeadline(time.Time{})
	}()
	go func() {
		defer close(cancelDone)
		select {
		case <-ctx.Done():
			_ = c.conn.SetDeadline(time.Now())
		case <-stopCancel:
		}
	}()

	if err := rpc.WriteRequest(c.conn, req); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if hasDeadline && time.Now().After(deadline) {
			return nil, context.DeadlineExceeded
		}
		return nil, err
	}

	resp, err := rpc.ReadResponse(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if hasDeadline && time.Now().After(deadline) {
			return nil, context.DeadlineExceeded
		}
		return nil, err
	}
	return resp, nil
}

func (c *rpcClient) Ping(ctx context.Context, id string) (*rpc.PingResponse, error) {
	resp, err := c.call(ctx, &rpc.RPCRequest{
		V:      1,
		Id:     id,
		Method: rpc.MethodPing,
		Params: &rpc.PingRequest{ClientSendTime: time.Now().UnixMilli()},
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("rpc: %w", rpc.Error{
			RequestID: resp.Id,
			Code:      resp.Error.Code,
			Message:   resp.Error.Message,
		})
	}
	if resp.Result == nil {
		return nil, fmt.Errorf("rpc: missing ping result")
	}
	return resp.Result, nil
}

func (c *rpcClient) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		closeErr = c.conn.Close()
	})
	return closeErr
}
