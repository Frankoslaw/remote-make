package domain

type MappingID string

type MatchType string

const (
	MatchTypeExact MatchType = "exact"
	MatchTypeGlob  MatchType = "glob"
)

type NodeSpecMapping struct {
	MappingID MappingID

	Match             map[string]string
	MatchType         MatchType
	ProviderOverrides map[ProviderID]map[string]any
}

func (m NodeSpecMapping) ID() MappingID {
	return m.MappingID
}
