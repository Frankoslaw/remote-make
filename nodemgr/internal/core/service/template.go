package service

import (
	"fmt"
	"maps"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"

	"github.com/gobwas/glob"
)

type TemplateService struct {
	templateRepository port.TemplateRepository
	mappingRepository  port.MappingRepository
}

func NewTemplateService(templateRepository port.TemplateRepository, mappingRepository port.MappingRepository) *TemplateService {
	return &TemplateService{templateRepository: templateRepository, mappingRepository: mappingRepository}
}

func (s *TemplateService) RenderTemplate(tmplID domain.TemplateID, providerID domain.ProviderID) (domain.NodeSpec, error) {
	if tmplID == "" {
		return domain.NodeSpec{}, fmt.Errorf("template id required")
	}

	tmpl, err := s.templateRepository.Get(tmplID)
	if err != nil {
		return domain.NodeSpec{}, fmt.Errorf("loading template: %w", err)
	}

	specified, err := s.applyTemplateOverrides(*tmpl, providerID)
	if err != nil {
		return domain.NodeSpec{}, err
	}

	resolved, err := s.ResolveSpec(specified)
	if err != nil {
		return domain.NodeSpec{}, err
	}

	return resolved, nil
}

func (s *TemplateService) ResolveSpec(spec domain.NodeSpec) (domain.NodeSpec, error) {
	mappings, err := s.mappingRepository.List()
	if err != nil {
		return domain.NodeSpec{}, fmt.Errorf("loading mappings: %w", err)
	}

	out := domain.NodeSpec{ProviderID: spec.ProviderID, Extra: map[string]any{}}
	maps.Copy(out.Extra, spec.Extra)

	for _, m := range mappings {
		if matchMapping(out, *m) {
			if vals, ok := m.ProviderOverrides["all"]; ok {
				maps.Copy(out.Extra, vals)
			}

			if vals, ok := m.ProviderOverrides[out.ProviderID]; ok {
				maps.Copy(out.Extra, vals)
			}
		}
	}

	return out, nil
}

func (s *TemplateService) applyTemplateOverrides(tmpl domain.NodeTemplate, providerID domain.ProviderID) (domain.NodeSpec, error) {
	out := domain.NodeSpec{
		ProviderID: providerID,
		Extra:      map[string]any{},
	}
	maps.Copy(out.Extra, tmpl.Extra)

	// canonical template fields
	out.Extra["name"] = tmpl.Name
	out.Extra["image"] = tmpl.Image
	out.Extra["user"] = tmpl.User

	out.Extra["cpus"] = tmpl.CPUs
	out.Extra["memory_mb"] = tmpl.MemoryMB
	out.Extra["disk_mb"] = tmpl.DiskMB

	if vals, ok := tmpl.ProviderOverrides["all"]; ok {
		maps.Copy(out.Extra, vals)
	}

	if vals, ok := tmpl.ProviderOverrides[providerID]; ok {
		maps.Copy(out.Extra, vals)
	}

	return out, nil
}

func matchMapping(spec domain.NodeSpec, mapping domain.TemplateMapping) bool {
	mtype := mapping.MatchType
	if mtype == "" {
		mtype = domain.MatchTypeExact
	}

	switch mtype {
	case domain.MatchTypeExact:
		for k, v := range mapping.Match {
			val, ok := spec.Extra[k]
			if !ok {
				return false
			}

			if fmt.Sprint(val) != v {
				return false
			}
		}
		return true

	case domain.MatchTypeGlob:
		for k, pat := range mapping.Match {
			val, ok := spec.Extra[k]
			if !ok {
				return false
			}
			str, ok := val.(string)
			if !ok {
				str = fmt.Sprint(val)
			}
			g, err := glob.Compile(pat)
			if err != nil {
				return false
			}
			if !g.Match(str) {
				return false
			}
		}
		return true

	default:
		return false
	}
}

var _ port.TemplateService = (*TemplateService)(nil)
