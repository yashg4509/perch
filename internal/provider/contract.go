package provider

import (
	"context"
	"errors"
)

// ErrUnsupported is returned when a provider definition does not implement an operation
// (e.g. Deploy on a read-only provider).
var ErrUnsupported = errors.New("provider: operation not supported")

// Provider is the contract every platform integration satisfies (spec: Provider interface).
type Provider interface {
	Logs(ctx context.Context, node Node, opts LogOpts) (<-chan LogLine, error)
	Status(ctx context.Context, node Node) (NodeStatus, error)
	Deploy(ctx context.Context, node Node) (DeployResult, error)
	Rollback(ctx context.Context, node Node, deploymentID string) error
	EnvList(ctx context.Context, node Node) ([]EnvVar, error)
	EnvSet(ctx context.Context, node Node, key, value string) error
	ListResources(ctx context.Context) ([]Resource, error)
}
