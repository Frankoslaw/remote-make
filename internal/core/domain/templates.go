package domain

import "github.com/google/uuid"

type WorkerTemplate struct {
	ID      uuid.UUID
	Name    string
	IsLocal bool
}

type TaskTemplate struct {
	ID             uuid.UUID
	Name           string
	IsAtomic       bool
	IsConcurrent   bool
	WorkerTemplate WorkerTemplate
	StepTemplates  []StepTemplate
}

type StepType int

const (
	UnknownStep StepType = 0
	ProcessStep          = iota + 1
	NotificationStep
	NestedTaskStep
)

type StepTemplate struct {
	ID       uuid.UUID
	SeqOrder int

	Type            StepType
	ProcessTemplate ProcessTemplate
	TaskTemplate    TaskTemplate
}

type ProcessTemplate struct {
	ID    uuid.UUID
	Cmd   string
	Pwd   string
	Stdin string
}
