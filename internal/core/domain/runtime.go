package domain

import "github.com/google/uuid"

type Worker struct {
	ID     uuid.UUID
	State  WorkerState
	NodeID uuid.UUID
}

type Task struct {
	ID     uuid.UUID
	State  TaskState
	Worker Worker
	Steps  []Step
}

type Step struct {
	ID    uuid.UUID
	State StepState

	ProcessResult Process
	Task          Task
}

type Process struct {
	ID    uuid.UUID
	State ProcessState

	ExitCode int
	Stdout   string
	Stderr   string
}
