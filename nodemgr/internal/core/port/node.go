package port

import (
	"github.com/google/uuid"

	"nodemgr/internal/core/domain"
)

type InfraBackend interface {
	Create() (*domain.Node, error)
	Destroy() error
	Conn() domain.NodeConn

	// optional for providers like docker or local
	Exec()
	Attach()
	Copy()
	Mount()
}

type NodeRepository interface {
	Save(*domain.Node) error
	Get(uuid.UUID) (*domain.Node, error)
	Delete(uuid.UUID) error
	List() ([]*domain.Node, error)
}

type NodeService interface {
	// Create a new node (creates domain object and persists it).
	Create(name string) (*domain.Node, error)

	// Get a node by ID.
	Get(id uuid.UUID) (*domain.Node, error)

	// Spawn causes the node to transition through spawn -> start -> running
	// (service implementation is responsible for calling provider APIs as needed).
	Spawn(id uuid.UUID) error

	// Terminate causes graceful termination (terminating -> finalized -> terminated).
	Terminate(id uuid.UUID) error

	// Interrupt marks the node interrupted (non-viable) and service should
	// ensure it ends up terminating.
	Interrupt(id uuid.UUID) error

	// Fail moves node to failed state and optionally records a reason.
	Fail(id uuid.UUID, reason string) error

	// List returns all nodes.
	List() ([]*domain.Node, error)
}
