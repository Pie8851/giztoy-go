package agenthost

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/genx"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/acl"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
)

var (
	ErrNilService       = errors.New("agenthost: nil service")
	ErrMissingHost      = errors.New("agenthost: host is required")
	ErrMissingPeerRun   = errors.New("agenthost: peer run store is required")
	ErrMissingSource    = errors.New("agenthost: stream source is required")
	ErrMissingConsumer  = errors.New("agenthost: stream consumer is required")
	ErrInvalidPublicKey = errors.New("agenthost: invalid public key")
)

type PeerRunStore interface {
	ResolveRunAgent(context.Context, giznet.PublicKey) (apitypes.AgentSelection, error)
	ActivateRunAgent(context.Context, giznet.PublicKey, apitypes.AgentSelection) (apitypes.PeerRunAgent, error)
}

type Authorizer interface {
	Authorize(context.Context, acl.AuthorizeRequest) error
}

type StreamSource interface {
	OpenAgentInput(context.Context) (genx.Stream, error)
}

type StreamSourceFunc func(context.Context) (genx.Stream, error)

func (f StreamSourceFunc) OpenAgentInput(ctx context.Context) (genx.Stream, error) {
	return f(ctx)
}

type StreamConsumer interface {
	ConsumeAgentOutput(context.Context, genx.Stream) error
}

type StreamConsumerFunc func(context.Context, genx.Stream) error

func (f StreamConsumerFunc) ConsumeAgentOutput(ctx context.Context, stream genx.Stream) error {
	return f(ctx, stream)
}

type Service struct {
	Host       genx.Transformer
	PeerRun    PeerRunStore
	Authorizer Authorizer
	PublicKey  giznet.PublicKey
	Source     StreamSource
	Consumer   StreamConsumer
	Logger     *slog.Logger
	Now        func() time.Time

	mu      sync.Mutex
	runtime *runtime
	status  apitypes.PeerRunStatus
}

func (s *Service) Reload(ctx context.Context) (apitypes.PeerRunStatus, error) {
	if s == nil {
		return apitypes.PeerRunStatus{}, ErrNilService
	}
	if err := s.validate(); err != nil {
		return s.setErrorStatus("", err), err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	selection, err := s.PeerRun.ResolveRunAgent(ctx, s.PublicKey)
	if err != nil {
		return s.setErrorStatus("", err), err
	}
	if err := s.authorize(ctx, selection); err != nil {
		return s.setErrorStatus(selection.WorkspaceName, err), err
	}
	s.setStatus(apitypes.PeerRunStatusStateStarting, selection.WorkspaceName, nil, nil)

	input, err := s.Source.OpenAgentInput(ctx)
	if err != nil {
		return s.setErrorStatus(selection.WorkspaceName, fmt.Errorf("agenthost: open input stream: %w", err)), err
	}
	if input == nil {
		err := errors.New("agenthost: input stream is required")
		return s.setErrorStatus(selection.WorkspaceName, err), err
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	output, err := s.Host.Transform(runCtx, workspacePattern(selection.WorkspaceName), input)
	if err != nil {
		cancel()
		_ = input.CloseWithError(err)
		return s.setErrorStatus(selection.WorkspaceName, err), err
	}
	if output == nil {
		cancel()
		_ = input.Close()
		err := errors.New("agenthost: output stream is required")
		return s.setErrorStatus(selection.WorkspaceName, err), err
	}
	if _, err := s.PeerRun.ActivateRunAgent(ctx, s.PublicKey, selection); err != nil {
		cancel()
		_ = errors.Join(output.CloseWithError(err), input.CloseWithError(err))
		return s.setErrorStatus(selection.WorkspaceName, err), err
	}

	now := s.now()
	next := &runtime{
		cancel:    cancel,
		input:     input,
		output:    output,
		done:      make(chan struct{}),
		workspace: selection.WorkspaceName,
		startedAt: now,
	}
	previous := s.swap(next)
	status := s.setStatus(apitypes.PeerRunStatusStateRunning, selection.WorkspaceName, nil, &now)
	go s.consume(runCtx, next)
	if err := previous.stop(ctx); err != nil {
		return s.setErrorStatus(selection.WorkspaceName, fmt.Errorf("agenthost: stop previous runtime: %w", err)), err
	}
	return status, nil
}

func (s *Service) Status(context.Context) (apitypes.PeerRunStatus, error) {
	if s == nil {
		return apitypes.PeerRunStatus{}, ErrNilService
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runtime != nil {
		return runningStatus(s.runtime.workspace, s.runtime.startedAt, s.now()), nil
	}
	if s.status.State == "" {
		return stoppedStatus(s.now()), nil
	}
	return s.status, nil
}

func (s *Service) Stop(ctx context.Context) (apitypes.PeerRunStatus, error) {
	if s == nil {
		return stoppedStatus(time.Now()), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	current := s.swap(nil)
	if current == nil {
		return s.setStatus(apitypes.PeerRunStatusStateStopped, "", nil, nil), nil
	}
	s.setStatus(apitypes.PeerRunStatusStateStopping, current.workspace, nil, &current.startedAt)
	if err := current.stop(ctx); err != nil {
		return s.setErrorStatus(current.workspace, err), err
	}
	return s.setStatus(apitypes.PeerRunStatusStateStopped, current.workspace, nil, nil), nil
}

func (s *Service) validate() error {
	switch {
	case s.Host == nil:
		return ErrMissingHost
	case s.PeerRun == nil:
		return ErrMissingPeerRun
	case s.PublicKey.IsZero():
		return ErrInvalidPublicKey
	case s.Source == nil:
		return ErrMissingSource
	case s.Consumer == nil:
		return ErrMissingConsumer
	default:
		return nil
	}
}

func (s *Service) authorize(ctx context.Context, selection apitypes.AgentSelection) error {
	if s.Authorizer == nil {
		return nil
	}
	return s.Authorizer.Authorize(ctx, acl.AuthorizeRequest{
		Subject:    acl.PublicKeySubject(s.PublicKey.String()),
		Resource:   acl.WorkspaceResource(selection.WorkspaceName),
		Permission: apitypes.ACLPermissionWorkspaceUse,
	})
}

func (s *Service) swap(next *runtime) *runtime {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.runtime
	s.runtime = next
	return previous
}

func (s *Service) consume(ctx context.Context, rt *runtime) {
	defer close(rt.done)
	err := s.Consumer.ConsumeAgentOutput(ctx, rt.output)
	if err != nil && ctx.Err() == nil {
		s.logger().Error("agenthost: output consumer failed", "error", err)
		s.mu.Lock()
		if s.runtime == rt {
			s.runtime = nil
		}
		s.mu.Unlock()
		s.setErrorStatus(rt.workspace, err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runtime == rt {
		s.runtime = nil
		s.status = stoppedStatus(s.now())
		if rt.workspace != "" {
			s.status.WorkspaceName = &rt.workspace
		}
	}
}

func (s *Service) setErrorStatus(workspace string, err error) apitypes.PeerRunStatus {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return s.setStatus(apitypes.PeerRunStatusStateError, workspace, &message, nil)
}

func (s *Service) setStatus(state apitypes.PeerRunStatusState, workspace string, message *string, startedAt *time.Time) apitypes.PeerRunStatus {
	now := s.now()
	status := apitypes.PeerRunStatus{
		State:     state,
		UpdatedAt: &now,
		Message:   message,
		StartedAt: startedAt,
	}
	if workspace != "" {
		status.WorkspaceName = &workspace
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
	return status
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) logger() *slog.Logger {
	if s != nil && s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

func workspacePattern(workspaceName string) string {
	return "workspaces/" + url.PathEscape(workspaceName)
}

func runningStatus(workspace string, startedAt, updatedAt time.Time) apitypes.PeerRunStatus {
	status := apitypes.PeerRunStatus{
		State:     apitypes.PeerRunStatusStateRunning,
		StartedAt: &startedAt,
		UpdatedAt: &updatedAt,
	}
	if workspace != "" {
		status.WorkspaceName = &workspace
	}
	return status
}

func stoppedStatus(updatedAt time.Time) apitypes.PeerRunStatus {
	return apitypes.PeerRunStatus{
		State:     apitypes.PeerRunStatusStateStopped,
		UpdatedAt: &updatedAt,
	}
}

type runtime struct {
	cancel    context.CancelFunc
	input     genx.Stream
	output    genx.Stream
	done      chan struct{}
	workspace string
	startedAt time.Time
}

func (r *runtime) stop(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.cancel()
	err := errors.Join(r.output.Close(), r.input.Close())
	select {
	case <-r.done:
		return err
	case <-ctx.Done():
		return errors.Join(err, ctx.Err())
	}
}
