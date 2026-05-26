package gizclaw

import (
	"context"
	"net"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

type rpcClient struct {
	peer *Client
}

func (c *rpcClient) Handle(conn net.Conn) error {
	return handleRPC(conn, c.dispatch)
}

func (c *rpcClient) dispatch(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if req == nil {
		return rpcapi.Error{Code: rpcapi.RPCErrorCodeInvalidRequest, Message: "nil request"}.RPCResponse(), nil
	}
	switch req.Method {
	case rpcapi.RPCMethodDeviceInfoGet:
		return c.handleGetDeviceInfo(ctx, req)
	case rpcapi.RPCMethodDeviceIdentifiersGet:
		return c.handleGetDeviceIdentifiers(ctx, req)
	case rpcapi.RPCMethodPeerPing:
		return handleRPCPing(ctx, req)
	default:
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeMethodNotFound, Message: "unsupported method: " + string(req.Method)}.RPCResponse(), nil
	}
}

func (c *rpcClient) Ping(ctx context.Context, conn net.Conn, id string) (*rpcapi.PingResponse, error) {
	return callRPCPing(ctx, conn, id)
}

func (c *rpcClient) GetDeviceInfo(ctx context.Context, conn net.Conn, id string) (*rpcapi.DeviceGetInfoResponse, error) {
	params, err := newRPCRequestParams(rpcapi.DeviceGetInfoRequest{}, (*rpcapi.RPCRequest_Params).FromDeviceGetInfoRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodDeviceInfoGet, params), rpcapi.RPCResponse_Result.AsDeviceGetInfoResponse)
	if err != nil {
		return nil, wrapRPCResultError("device info", err)
	}
	return result, nil
}

func (c *rpcClient) GetDeviceIdentifiers(ctx context.Context, conn net.Conn, id string) (*rpcapi.DeviceGetIdentifiersResponse, error) {
	params, err := newRPCRequestParams(rpcapi.DeviceGetIdentifiersRequest{}, (*rpcapi.RPCRequest_Params).FromDeviceGetIdentifiersRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodDeviceIdentifiersGet, params), rpcapi.RPCResponse_Result.AsDeviceGetIdentifiersResponse)
	if err != nil {
		return nil, wrapRPCResultError("device identifiers", err)
	}
	return result, nil
}

func (c *rpcClient) GetServerInfo(ctx context.Context, conn net.Conn, id string) (*rpcapi.ServerGetInfoResponse, error) {
	params, err := newRPCRequestParams(rpcapi.ServerGetInfoRequest{}, (*rpcapi.RPCRequest_Params).FromServerGetInfoRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodServerInfoGet, params), rpcapi.RPCResponse_Result.AsServerGetInfoResponse)
	if err != nil {
		return nil, wrapRPCResultError("server info", err)
	}
	return result, nil
}

func (c *rpcClient) handleGetDeviceInfo(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if err := validateRPCParams(req.Params, rpcapi.RPCRequest_Params.AsDeviceGetInfoRequest); err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.peer == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "peer client not configured"}.RPCResponse(), nil
	}
	result, err := convertRPCType[rpcapi.DeviceGetInfoResponse](gearDeviceToPeerRefreshInfo(c.peer.Device))
	if err != nil {
		return nil, err
	}
	return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromDeviceGetInfoResponse)
}

func (c *rpcClient) handleGetDeviceIdentifiers(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if err := validateRPCParams(req.Params, rpcapi.RPCRequest_Params.AsDeviceGetIdentifiersRequest); err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.peer == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "peer client not configured"}.RPCResponse(), nil
	}
	result, err := convertRPCType[rpcapi.DeviceGetIdentifiersResponse](gearDeviceToPeerRefreshIdentifiers(c.peer.Device))
	if err != nil {
		return nil, err
	}
	return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromDeviceGetIdentifiersResponse)
}

func (c *rpcClient) GetPeerInfo(ctx context.Context, conn net.Conn, id string) (*rpcapi.PeerGetInfoResponse, error) {
	params, err := newRPCRequestParams(rpcapi.PeerGetInfoRequest{}, (*rpcapi.RPCRequest_Params).FromPeerGetInfoRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodPeerInfoGet, params), rpcapi.RPCResponse_Result.AsPeerGetInfoResponse)
	if err != nil {
		return nil, wrapRPCResultError("peer info", err)
	}
	return result, nil
}

func (c *rpcClient) PutPeerInfo(ctx context.Context, conn net.Conn, id string, info rpcapi.PeerPutInfoRequest) (*rpcapi.PeerPutInfoResponse, error) {
	params, err := newRPCRequestParams(info, (*rpcapi.RPCRequest_Params).FromPeerPutInfoRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodPeerInfoPut, params), rpcapi.RPCResponse_Result.AsPeerPutInfoResponse)
	if err != nil {
		return nil, wrapRPCResultError("peer info", err)
	}
	return result, nil
}

func (c *rpcClient) GetPeerRuntime(ctx context.Context, conn net.Conn, id string) (*rpcapi.PeerGetRuntimeResponse, error) {
	params, err := newRPCRequestParams(rpcapi.PeerGetRuntimeRequest{}, (*rpcapi.RPCRequest_Params).FromPeerGetRuntimeRequest)
	if err != nil {
		return nil, err
	}
	result, err := callRPCResult(ctx, conn, newRPCRequest(id, rpcapi.RPCMethodPeerRuntimeGet, params), rpcapi.RPCResponse_Result.AsPeerGetRuntimeResponse)
	if err != nil {
		return nil, wrapRPCResultError("peer runtime", err)
	}
	return result, nil
}
