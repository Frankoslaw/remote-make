package port

import (
	"context"
	"nodemgr/internal/core/domain"
)

type NodeClient interface {
	ID() string

	// TODO: Avoid using pointers for this s
	Conn() (*domain.NodeConn, error)

	Tag(key string) (string, bool)
	SetTag(key, val string)

	HasCap(key string) bool
	SetCap(key string, val bool)
}

type Provisioner interface {
	ID() string
	Up(ctx context.Context, spec map[string]string) (NodeClient, error)
	Destroy(ctx context.Context, nodeID string) error
}
