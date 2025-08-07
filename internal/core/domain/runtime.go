package domain

import (
	"github.com/google/uuid"
	"github.com/looplab/fsm"
)

type Worker struct {
	ID    uuid.UUID
	State *fsm.FSM
	Err   error
	Tmpl  *WorkerTemplate

	NodeID uuid.UUID
}

func NewWorker(wt *WorkerTemplate) Worker {
	w := Worker{
		ID:   uuid.New(),
		Tmpl: wt,
	}

	w.State = fsm.NewFSM(
		"scheduled",
		fsm.Events{
			{Name: "provision", Src: []string{"scheduled"}, Dst: "provisioning"},
			{Name: "provisioned", Src: []string{"provisioning"}, Dst: "provisioned"},
			{Name: "terminate", Src: []string{"provisioned"}, Dst: "terminating"},
			{Name: "terminated", Src: []string{"terminating"}, Dst: "terminated"},
			{Name: "error", Src: []string{"provisioning", "terminating"}, Dst: "error"},
		},
		fsm.Callbacks{},
	)

	return w
}

type Task struct {
	ID    uuid.UUID
	State *fsm.FSM
	Err   error
	Tmpl  *TaskTemplate

	Worker Worker
	Steps  []Step
}

func NewTask(tt *TaskTemplate) Task {
	t := Task{
		ID:   uuid.New(),
		Tmpl: tt,
	}

	t.State = fsm.NewFSM(
		"scheduled",
		fsm.Events{
			{Name: "start", Src: []string{"scheduled"}, Dst: "running"},
			{Name: "complete", Src: []string{"running"}, Dst: "done"},
			{Name: "error", Src: []string{"running"}, Dst: "error"},
		},
		fsm.Callbacks{},
	)

	return t
}

type Step struct {
	ID    uuid.UUID
	State *fsm.FSM
	Err   error
	Tmpl  *StepTemplate

	ProcessResult *Process
	TaskResult    *Task
}

func NewStep(st *StepTemplate) Step {
	s := Step{
		ID:   uuid.New(),
		Tmpl: st,
	}

	s.State = fsm.NewFSM(
		"scheduled",
		fsm.Events{
			{Name: "start", Src: []string{"scheduled"}, Dst: "running"},
			{Name: "complete", Src: []string{"running"}, Dst: "done"},
			{Name: "error", Src: []string{"running"}, Dst: "error"},
		},
		fsm.Callbacks{},
	)

	return s
}

type Process struct {
	ID    uuid.UUID
	State *fsm.FSM
	Err   error
	Tmpl  *ProcessTemplate

	ExitCode int
	Stdout   string
	Stderr   string
}

func NewProcess(pt *ProcessTemplate) Process {
	p := Process{
		ID:   uuid.New(),
		Tmpl: pt,
	}

	p.State = fsm.NewFSM(
		"scheduled",
		fsm.Events{
			{Name: "start", Src: []string{"scheduled"}, Dst: "running"},
			{Name: "complete", Src: []string{"running"}, Dst: "done"},
			{Name: "error", Src: []string{"running"}, Dst: "error"},
		},
		fsm.Callbacks{},
	)

	return p
}
