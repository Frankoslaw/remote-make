package port

import "nodemgr/internal/core/domain"

type TemplateRepository interface {
	Create(tmpl domain.NodeTemplate) error
	Get(id domain.TemplateID) (*domain.NodeTemplate, error)
	List() ([]*domain.NodeTemplate, error)
	Delete(id domain.TemplateID) error
}

type TemplateService interface {
	CreateTemplate(template domain.NodeTemplate) (domain.TemplateID, error)
	GetTemplate(id domain.TemplateID) (*domain.NodeTemplate, error)
	ListTemplates() ([]*domain.NodeTemplate, error)
	DeleteTemplate(id domain.TemplateID) error

	RenderTemplate(templateID domain.TemplateID, providerID domain.ProviderID) (domain.NodeSpec, error)
}
