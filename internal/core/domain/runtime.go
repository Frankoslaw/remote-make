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

	ProcessResult ProcessResult
	Task          Task
}

type ProcessResult struct {
	ID       uuid.UUID
	ExitCode int
	Stdout   string
	Stderr   string
}
