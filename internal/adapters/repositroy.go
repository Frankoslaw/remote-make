package adapters

import (
	"errors"
	"remote-make/internal/core/domain"

	"github.com/google/uuid"
)

type InMemoryTemplateRepo struct {
	templates map[uuid.UUID]domain.TaskTemplate
}

func NewInMemoryTemplateRepo(tmpls []domain.TaskTemplate) *InMemoryTemplateRepo {
	t := map[uuid.UUID]domain.TaskTemplate{}
	for _, tmpl := range tmpls {
		t[tmpl.ID] = tmpl
	}
	return &InMemoryTemplateRepo{templates: t}
}

func (r *InMemoryTemplateRepo) GetTaskTemplate(id uuid.UUID) (domain.TaskTemplate, error) {
	if tmpl, ok := r.templates[id]; ok {
		return tmpl, nil
	}
	return domain.TaskTemplate{}, errors.New("template not found")
}

func (r *InMemoryTemplateRepo) ListTaskTemplates() ([]domain.TaskTemplate, error) {
	var list []domain.TaskTemplate
	for _, t := range r.templates {
		list = append(list, t)
	}
	return list, nil
}
