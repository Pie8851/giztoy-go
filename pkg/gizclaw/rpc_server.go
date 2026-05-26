package gizclaw

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/serverpublic"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/peer"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
)

type rpcServerInfoService interface {
	GetServerInfo(context.Context, serverpublic.GetServerInfoRequestObject) (serverpublic.GetServerInfoResponseObject, error)
}

type rpcPeerService interface {
	GetSelfInfo(context.Context, giznet.PublicKey) (apitypes.DeviceInfo, error)
	PutSelfInfo(context.Context, giznet.PublicKey, apitypes.DeviceInfo) (apitypes.DeviceInfo, error)
	GetSelfRuntime(context.Context, giznet.PublicKey) apitypes.Runtime
}

type rpcServer struct {
	peer            rpcPeerService
	serverInfo      rpcServerInfoService
	callerPublicKey giznet.PublicKey
}

func (s *rpcServer) Ping(ctx context.Context, conn net.Conn, id string) (*rpcapi.PingResponse, error) {
	return callRPCPing(ctx, conn, id)
}

func (s *rpcServer) Handle(conn net.Conn) error {
	return handleRPC(conn, s.dispatch)
}

func (s *rpcServer) dispatch(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if req == nil {
		return rpcapi.Error{Code: rpcapi.RPCErrorCodeInvalidRequest, Message: "nil request"}.RPCResponse(), nil
	}
	switch req.Method {
	case rpcapi.RPCMethodPeerPing:
		return handleRPCPing(ctx, req)
	case rpcapi.RPCMethodServerInfoGet:
		return s.handleGetServerInfo(ctx, req)
	case rpcapi.RPCMethodPeerInfoGet:
		return s.handleGetInfo(ctx, req)
	case rpcapi.RPCMethodPeerInfoPut:
		return s.handlePutInfo(ctx, req)
	case rpcapi.RPCMethodPeerRuntimeGet:
		return s.handleGetRuntime(ctx, req)
	default:
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeMethodNotFound, Message: fmt.Sprintf("unknown method: %s", req.Method)}.RPCResponse(), nil
	}
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

func (s *rpcServer) handleGetInfo(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if err := validateRPCParams(req.Params, rpcapi.RPCRequest_Params.AsPeerGetInfoRequest); err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	if s.peer == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "peer service not configured"}.RPCResponse(), nil
	}
	resp, err := s.peer.GetSelfInfo(ctx, s.callerPublicKey)
	if err != nil {
		if errors.Is(err, peer.ErrPeerNotFound) {
			return rpcAPIError(req.Id, http.StatusNotFound, apitypes.NewErrorResponse("PEER_NOT_FOUND", err.Error())), nil
		}
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: err.Error()}.RPCResponse(), nil
	}
	result, err := convertRPCType[rpcapi.PeerGetInfoResponse](resp)
	if err != nil {
		return nil, err
	}
	return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromPeerGetInfoResponse)
}

func (s *rpcServer) handlePutInfo(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if req.Params == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInvalidParams, Message: "missing params"}.RPCResponse(), nil
	}
	params, err := req.Params.AsPeerPutInfoRequest()
	if err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	body, err := convertRPCType[apitypes.DeviceInfo](params)
	if err != nil {
		return nil, err
	}
	if s.peer == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "peer service not configured"}.RPCResponse(), nil
	}
	resp, err := s.peer.PutSelfInfo(ctx, s.callerPublicKey, body)
	if err != nil {
		if errors.Is(err, peer.ErrPeerNotFound) {
			return rpcAPIError(req.Id, http.StatusNotFound, apitypes.NewErrorResponse("PEER_NOT_FOUND", err.Error())), nil
		}
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: err.Error()}.RPCResponse(), nil
	}
	result, err := convertRPCType[rpcapi.PeerPutInfoResponse](resp)
	if err != nil {
		return nil, err
	}
	return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromPeerPutInfoResponse)
}

func (s *rpcServer) handleGetRuntime(ctx context.Context, req *rpcapi.RPCRequest) (*rpcapi.RPCResponse, error) {
	if err := validateRPCParams(req.Params, rpcapi.RPCRequest_Params.AsPeerGetRuntimeRequest); err != nil {
		return rpcInvalidParams(req.Id), nil
	}
	if s.peer == nil {
		return rpcapi.Error{RequestID: req.Id, Code: rpcapi.RPCErrorCodeInternalError, Message: "peer service not configured"}.RPCResponse(), nil
	}
	result, err := convertRPCType[rpcapi.PeerGetRuntimeResponse](s.peer.GetSelfRuntime(ctx, s.callerPublicKey))
	if err != nil {
		return nil, err
	}
	return newRPCResultResponse(req.Id, result, (*rpcapi.RPCResponse_Result).FromPeerGetRuntimeResponse)
}

func (s *rpcServer) serverInfoContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return serverpublic.WithCallerPublicKey(ctx, s.callerPublicKey)
}
