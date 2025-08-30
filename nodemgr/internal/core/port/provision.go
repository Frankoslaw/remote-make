package port

import "nodemgr/internal/core/domain"

// internal providers replacable for each backend
type NodeRepository interface {
	Create(node domain.Node) error
	Get(id domain.NodeID) (*domain.Node, error)
	List() ([]*domain.Node, error)
	Delete(id domain.NodeID) error
}
type ProviderRepository interface {
	Create(provider NodeProvider) error
	Get(id domain.ProviderID) (*NodeProvider, error)
	List() []*NodeProvider
	Delete(id domain.ProviderID) error
}

type NodeProvider interface {
	ID() domain.ProviderID
	Provision(spec domain.NodeSpec) (domain.Node, error)
	Destroy(nodeID domain.NodeID) error

	// optional
	Controller(node *domain.Node) (NodeController, error)
}
type NodeController interface {
	Start() error
	Stop() error
	Reboot() error
	Hibernate() error
	Terminate() error
}

// user facing apis
type ProvisionService interface {
	Provision(spec domain.NodeSpec) (*domain.Node, error)
	Destroy(nodeID domain.NodeID) error
}
