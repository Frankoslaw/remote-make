package service

import (
	"fmt"
	"maps"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/util"
)

type TemplateService struct {
	templateRepository port.TemplateRepository
}

func NewTemplateService(templateRepository port.TemplateRepository) *TemplateService {
	return &TemplateService{templateRepository: templateRepository}
}

func (s *TemplateService) CreateTemplate(tmpl domain.NodeTemplate) (domain.TemplateID, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *TemplateService) GetTemplate(tmplID domain.TemplateID) (*domain.NodeTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *TemplateService) ListTemplates() ([]*domain.NodeTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *TemplateService) DeleteTemplate(tmplID domain.TemplateID) error {
	return fmt.Errorf("not implemented")
}

func (s *TemplateService) RenderTemplate(templateID domain.TemplateID, providerID domain.ProviderID) (domain.NodeSpec, error) {
	tmpl, err := s.templateRepository.Get(templateID)
	if err != nil {
		return domain.NodeSpec{}, fmt.Errorf("loading template: %w", err)
	}

	out := domain.NodeSpec{
		ProviderID: providerID,
		Extra:      map[string]any{},
	}
	maps.Copy(out.Extra, tmpl.Extra)

	extra, err := util.StructToMapJSON(tmpl)
	if err != nil {
		return domain.NodeSpec{}, err
	}
	out.Extra = extra

	if vals, ok := tmpl.ProviderOverrides["all"]; ok {
		maps.Copy(out.Extra, vals)
	}

	if vals, ok := tmpl.ProviderOverrides[providerID]; ok {
		maps.Copy(out.Extra, vals)
	}

	return out, nil
}

var _ port.TemplateService = (*TemplateService)(nil)
