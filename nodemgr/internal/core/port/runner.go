package port

import (
	"context"
	"io"
	"nodemgr/internal/core/domain"
)

type Runner interface {
	ID() string

	Copy(ctx context.Context, src, dst string) error
	Attach(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer) (wait func() error, err error)
	Exec(ctx context.Context, cmd domain.Command) (domain.ExecResult, error)
}

// Special case of runner which alters the capabilities of the node by configuring additional software
type Bootstrapper interface {
	ID() string
	Configure(ctx context.Context, node NodeClient) (NodeClient, error)
}
