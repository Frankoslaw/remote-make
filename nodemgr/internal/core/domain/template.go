package domain

type TemplateID string

type ImageType string

const (
	ImageTypeAlias  ImageType = "alias"
	ImageTypeDocker ImageType = "docker"
	ImageTypeISO    ImageType = "iso"
	ImageTypeQCOW2  ImageType = "qcow2"
	ImageTypeAMI    ImageType = "ami"
)

type NodeTemplate struct {
	TemplateID TemplateID `json:"-"`

	Name      string    `json:"name"`
	Image     string    `json:"image"`
	ImageType ImageType `json:"image_type"`
	User      string    `json:"user"`
	CPUs      int       `json:"cpus"`
	MemoryMB  int       `json:"memory_mb"`
	DiskMB    int       `json:"disk_mb"`

	Extra             map[string]any                `json:"-"`
	ProviderOverrides map[ProviderID]map[string]any `json:"-"`
}

func (n NodeTemplate) ID() TemplateID {
	return n.TemplateID
}
