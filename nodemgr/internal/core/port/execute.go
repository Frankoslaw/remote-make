package port

import (
	"io"
	"nodemgr/internal/core/domain"
)

type NodeExecProviderRepository interface {
	Create(provider NodeExecProvider) error
	Get(id domain.ExecProviderID) (*NodeExecProvider, error)
	List() ([]*NodeExecProvider, error)
	Delete(id domain.ExecProviderID) error
}
type ExecHandleRepository interface {
	Create(handle ExecHandle) error
	Get(id domain.ExecHandleID) (*ExecHandle, error)
	List() ([]*ExecHandle, error)
	Delete(id domain.ExecHandleID) error
}

type NodeExecProvider interface {
	ID() domain.ExecProviderID
	OpenExecHandle(node *domain.Node) (ExecHandle, error)
}

type ExecHandle interface {
	ID() domain.ExecHandleID
	Close() error

	Attach(attach domain.AttachRequest) (*domain.AttachResult, error)
	Exec(req domain.ExecRequest) (*domain.ExecResult, error)
	ExecStream(exec domain.ExecRequest, attach domain.AttachRequest) (*domain.AttachResult, error)

	CopyTo(src io.Reader, dst string) error
	CopyFrom(src string) (io.ReadCloser, error)
}

type NodeExecuteService interface {
	Attach(req domain.AttachRequest) (*domain.AttachResult, error)
	Exec(req domain.ExecRequest) (*domain.ExecResult, error)
	ExecStream(exec domain.ExecRequest, attach domain.AttachRequest) (*domain.AttachResult, error)

	CopyTo(req domain.CopyToRequest) error
	CopyFrom(req domain.CopyFromRequest) error
}
