package service

import (
	"fmt"
	"maps"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"

	"github.com/gobwas/glob"
)

type MappingService struct {
	mappingRepository port.MappingRepository
}

func NewMappingService(mappingRepository port.MappingRepository) *MappingService {
	return &MappingService{mappingRepository: mappingRepository}
}

func (s *MappingService) CreateMapping(mapping domain.NodeSpecMapping) (domain.MappingID, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *MappingService) GetMapping(mappingID domain.MappingID) (*domain.NodeSpecMapping, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MappingService) ListMappings() ([]*domain.NodeSpecMapping, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MappingService) DeleteMapping(mappingID domain.MappingID) error {
	return fmt.Errorf("not implemented")
}

func (s *MappingService) ResolveSpecAliases(spec domain.NodeSpec) (domain.NodeSpec, error) {
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

func matchMapping(spec domain.NodeSpec, mapping domain.NodeSpecMapping) bool {
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

var _ port.MappingService = (*MappingService)(nil)
