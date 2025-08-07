package domain

import "github.com/google/uuid"

type WorkerTemplate struct {
	ID uuid.UUID

	Backend     string
	DockerImage string
}

type TaskTemplate struct {
	ID uuid.UUID

	IsAtomic       bool
	IsConcurrent   bool
	WorkerTemplate WorkerTemplate
	StepTemplates  []StepTemplate
}

type StepKind int

const (
	StepKindProcess StepKind = iota
	StepKindTask
)

type StepTemplate struct {
	ID       uuid.UUID
	SeqOrder int

	Kind            StepKind
	ProcessTemplate *ProcessTemplate
	TaskTemplate    *TaskTemplate
}

type ProcessTemplate struct {
	ID uuid.UUID

	Cmd   string
	Pwd   string
	Stdin string
}
