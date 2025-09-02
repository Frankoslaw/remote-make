package port

import "nodemgr/internal/core/domain"

type NodeRepository interface {
	Create(node domain.Node) error
	Get(id domain.NodeID) (*domain.Node, error)
	List() ([]*domain.Node, error)
	Delete(id domain.NodeID) error
}

type NodeProviderRepository interface {
	Create(provider NodeProvider) error
	Get(id domain.ProviderID) (*NodeProvider, error)
	List() []*NodeProvider
	Delete(id domain.ProviderID) error
}

type NodeProvider interface {
	ID() domain.ProviderID
	Provision(spec domain.NodeSpec) (*domain.Node, error)
	Destroy(nodeID domain.NodeID) error
}

type NodeProvisionService interface {
	ProvisionNode(spec domain.NodeSpec) (*domain.Node, error)
	DestroyNode(nodeID domain.NodeID) error
}
