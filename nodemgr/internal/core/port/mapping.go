package port

import "nodemgr/internal/core/domain"

type MappingRepository interface {
	Create(mapping domain.NodeSpecMapping) error
	Get(id domain.MappingID) (*domain.NodeSpecMapping, error)
	List() ([]*domain.NodeSpecMapping, error)
	Delete(id domain.MappingID) error
}

type MappingService interface {
	CreateMapping(mapping domain.NodeSpecMapping) (domain.MappingID, error)
	GetMapping(id domain.MappingID) (*domain.NodeSpecMapping, error)
	ListMappings() ([]*domain.NodeSpecMapping, error)
	DeleteMapping(id domain.MappingID) error

	ResolveSpecAliases(spec domain.NodeSpec) (domain.NodeSpec, error)
}
