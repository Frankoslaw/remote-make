package ports

import (
	"remote-make/internal/core/domain"

	"github.com/google/uuid"
)

type TemplateRepo interface {
	GetTaskTemplate(id uuid.UUID) (domain.TaskTemplate, error)
	ListTaskTemplates() ([]domain.TaskTemplate, error)
}

type ProcessRunner interface {
	Run(pt domain.ProcessTemplate) (domain.ProcessResult, error)
}
