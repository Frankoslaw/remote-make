package service

import (
	"fmt"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
)

type ProvisionService struct {
	nodeRepository     port.NodeRepository
	providerRepository port.ProviderRepository
}

func NewProvisionService(nodeRepository port.NodeRepository, providerRepository port.ProviderRepository) *ProvisionService {
	return &ProvisionService{nodeRepository: nodeRepository, providerRepository: providerRepository}
}

func (s *ProvisionService) Provision(spec domain.NodeSpec) (*domain.Node, error) {
	// TODO
	return nil, fmt.Errorf("TODO")
}

func (s *ProvisionService) Destroy(nodeID domain.NodeID) error {
	// TODO
	return fmt.Errorf("TODO")
}

var _ port.ProvisionService = (*ProvisionService)(nil)
