package domain

type TemplateID string
type MappingID string
type MatchType string

const (
	MatchTypeExact MatchType = "exact"
	MatchTypeGlob  MatchType = "glob"
)

type NodeTemplate struct {
	TemplateID TemplateID

	Name  string
	Image string
	User  string

	CPUs     int // vCPUs
	MemoryMB int
	DiskMB   int

	Extra             map[string]any
	ProviderOverrides map[ProviderID]map[string]any
}

func (n NodeTemplate) ID() TemplateID {
	return n.TemplateID
}

type TemplateMapping struct {
	MappingID MappingID

	Match             map[string]string
	MatchType         MatchType // glob | exact
	ProviderOverrides map[ProviderID]map[string]any
}

func (n TemplateMapping) ID() MappingID {
	return n.MappingID
}
