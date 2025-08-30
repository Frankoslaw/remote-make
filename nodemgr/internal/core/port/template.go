package port

import "nodemgr/internal/core/domain"

type TemplateRepository interface {
	Create(tmpl domain.NodeTemplate) error
	Get(id domain.TemplateID) (*domain.NodeTemplate, error)
	List() ([]*domain.NodeTemplate, error)
	Delete(id domain.TemplateID) error
}

type MappingRepository interface {
	Create(mapping domain.TemplateMapping) error
	Get(id domain.MappingID) (*domain.TemplateMapping, error)
	List() ([]*domain.TemplateMapping, error)
	Delete(id domain.MappingID) error
}

type TemplateService interface {
	RenderTemplate(tmplID domain.TemplateID, providerID domain.ProviderID) (domain.NodeSpec, error)
	ResolveSpec(spec domain.NodeSpec) (domain.NodeSpec, error)
}
