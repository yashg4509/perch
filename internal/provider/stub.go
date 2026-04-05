package provider

import (
	"context"
	"fmt"
)

// ReadOnlyStub is a test-oriented [Provider] backed by static JSON and optional resource lists.
type ReadOnlyStub struct {
	spec *Spec
	opts ReadOnlyStubOptions
}

// ReadOnlyStubOptions configures canned responses for [ReadOnlyStub].
type ReadOnlyStubOptions struct {
	StatusBody []byte
	Resources  []Resource
}

// NewReadOnlyStub returns a [Provider] that implements read-only operations for tests.
func NewReadOnlyStub(spec *Spec, opts ReadOnlyStubOptions) *ReadOnlyStub {
	return &ReadOnlyStub{spec: spec, opts: opts}
}

func (s *ReadOnlyStub) Logs(ctx context.Context, node Node, opts LogOpts) (<-chan LogLine, error) {
	ch := make(chan LogLine)
	close(ch)
	return ch, nil
}

func (s *ReadOnlyStub) Status(ctx context.Context, node Node) (NodeStatus, error) {
	_ = ctx
	_ = node
	_ = s.spec
	raw := s.opts.StatusBody
	if len(raw) == 0 {
		raw = []byte(`{"healthy":false}`)
	}
	var st NodeStatus
	if err := DecodeJSON(raw, &st); err != nil {
		return NodeStatus{}, fmt.Errorf("read-only stub status: %w", err)
	}
	return st, nil
}

func (s *ReadOnlyStub) Deploy(ctx context.Context, node Node) (DeployResult, error) {
	return DeployResult{}, ErrUnsupported
}

func (s *ReadOnlyStub) Rollback(ctx context.Context, node Node, deploymentID string) error {
	return ErrUnsupported
}

func (s *ReadOnlyStub) EnvList(ctx context.Context, node Node) ([]EnvVar, error) {
	return nil, ErrUnsupported
}

func (s *ReadOnlyStub) EnvSet(ctx context.Context, node Node, key, value string) error {
	return ErrUnsupported
}

func (s *ReadOnlyStub) ListResources(ctx context.Context) ([]Resource, error) {
	return s.opts.Resources, nil
}

var _ Provider = (*ReadOnlyStub)(nil)
