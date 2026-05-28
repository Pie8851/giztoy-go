package gizclaw

import (
	"context"
	"net"
	"net/http"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/serverpublic"
)

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

func (s *rpcServer) handleGetServerInfo(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if err := validateRPCParams(req.Params, rpcapi.RPCRequest_Params.AsServerGetInfoRequest); err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	if s.serverInfo == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "server info service not configured"}.RPCResponse(), nil
	}
	resp, err := s.serverInfo.GetServerInfo(s.serverInfoContext(ctx), serverpublic.GetServerInfoRequestObject{})
	if err != nil {
		return nil, err
	}
	switch body := resp.(type) {
	case serverpublic.GetServerInfo200JSONResponse:
		result, err := convertRPCType[rpcapi.ServerGetInfoResponse](body)
		if err != nil {
			return nil, err
		}
		return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromServerGetInfoResponse)
	case serverpublic.GetServerInfo400JSONResponse:
		return rpcAPIError(req.Id, http.StatusBadRequest, apitypes.ErrorResponse(body)), nil
	default:
		return rpcUnexpectedResponse(req.Id, resp), nil
	}
}

func (s *rpcServer) serverInfoContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return serverpublic.WithCallerPublicKey(ctx, s.callerPublicKey)
}
