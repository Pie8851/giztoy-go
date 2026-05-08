package rpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestFrameRequestResponseRoundTrip(t *testing.T) {
	var frameBuf bytes.Buffer
	if err := WriteFrame(&frameBuf, []byte("payload")); err != nil {
		t.Fatalf("WriteFrame() error = %v", err)
	}
	frame, err := ReadFrame(&frameBuf)
	if err != nil {
		t.Fatalf("ReadFrame() error = %v", err)
	}
	if string(frame) != "payload" {
		t.Fatalf("ReadFrame() = %q", frame)
	}

	var reqBuf bytes.Buffer
	req := &RPCRequest{
		V:      1,
		Id:     "req-1",
		Method: MethodPing,
		Params: &PingRequest{ClientSendTime: 123},
	}
	if err := WriteRequest(&reqBuf, req); err != nil {
		t.Fatalf("WriteRequest() error = %v", err)
	}
	gotReq, err := ReadRequest(&reqBuf)
	if err != nil {
		t.Fatalf("ReadRequest() error = %v", err)
	}
	if gotReq.Id != req.Id || gotReq.Method != MethodPing || gotReq.Params == nil || gotReq.Params.ClientSendTime != 123 {
		t.Fatalf("ReadRequest() = %+v", gotReq)
	}

	var respBuf bytes.Buffer
	resp := &RPCResponse{
		V:      1,
		Id:     "req-1",
		Result: &PingResponse{ServerTime: 456},
	}
	if err := WriteResponse(&respBuf, resp); err != nil {
		t.Fatalf("WriteResponse() error = %v", err)
	}
	gotResp, err := ReadResponse(&respBuf)
	if err != nil {
		t.Fatalf("ReadResponse() error = %v", err)
	}
	if gotResp.Id != resp.Id || gotResp.Result == nil || gotResp.Result.ServerTime != 456 {
		t.Fatalf("ReadResponse() = %+v", gotResp)
	}
}

func TestWriteFramePropagatesHeaderWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	if err := WriteFrame(errorWriter{err: writeErr}, []byte("payload")); !errors.Is(err, writeErr) {
		t.Fatalf("WriteFrame() err = %v, want %v", err, writeErr)
	}
}

func TestReadRequestAndResponseRejectInvalidJSON(t *testing.T) {
	var reqBuf bytes.Buffer
	if err := WriteFrame(&reqBuf, []byte("{")); err != nil {
		t.Fatalf("WriteFrame(request) error = %v", err)
	}
	if _, err := ReadRequest(&reqBuf); err == nil {
		t.Fatal("ReadRequest() should fail for invalid JSON")
	}

	var respBuf bytes.Buffer
	if err := WriteFrame(&respBuf, []byte("{")); err != nil {
		t.Fatalf("WriteFrame(response) error = %v", err)
	}
	if _, err := ReadResponse(&respBuf); err == nil {
		t.Fatal("ReadResponse() should fail for invalid JSON")
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestReadFrameRejectsOversizedFrame(t *testing.T) {
	var buf bytes.Buffer
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], MaxFrameSize+1)
	if _, err := buf.Write(hdr[:]); err != nil {
		t.Fatalf("Write(header) error = %v", err)
	}

	_, err := ReadFrame(&buf)
	if err == nil || err.Error() != "rpc: frame too large: 1048577" {
		t.Fatalf("ReadFrame() err = %v", err)
	}
}

func TestErrorImplementsErrorAndBuildsRPCResponse(t *testing.T) {
	rpcErr := Error{RequestID: "req-1", Code: -32602, Message: "missing params"}
	var err error = rpcErr
	if err.Error() != "missing params" {
		t.Fatalf("Error() = %q", err.Error())
	}

	errResp := rpcErr.RPCResponse()
	if errResp.V != 1 || errResp.Id != "req-1" || errResp.Error == nil {
		t.Fatalf("RPCResponse() = %+v", errResp)
	}
	if errResp.Error.Code != -32602 || errResp.Error.Message != "missing params" {
		t.Fatalf("RPCResponse().Error = %+v", errResp.Error)
	}
}

func TestErrorUsesFallbackMessage(t *testing.T) {
	rpcErr := Error{Code: -1}
	if rpcErr.Error() != "rpc error -1" {
		t.Fatalf("Error() = %q", rpcErr.Error())
	}

	errResp := rpcErr.RPCResponse()
	if errResp.Error == nil || errResp.Error.Message != "rpc error -1" {
		t.Fatalf("RPCResponse().Error = %+v", errResp.Error)
	}
}
