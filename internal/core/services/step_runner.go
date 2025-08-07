package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
)

type StepRunner struct {
	nodeIDRepo ports.NodeIdentityRepo
	eventBus   ports.EventBus
	procRunner ports.ProcessRunner
}

func NewStepRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, pr ports.ProcessRunner) *StepRunner {
	return &StepRunner{nodeIDRepo: ni, eventBus: ev, procRunner: pr}
}

func (s *StepRunner) Start(ctx context.Context, step domain.Step) (domain.Step, error) {
	slog.Debug("Starting step", "step_id", step.ID)
	step.State.Event(ctx, "start")

	if step.Tmpl.Kind == domain.StepKindTask {
		var err error
		task := domain.NewTask(step.Tmpl.TaskTemplate)
		task, err = s.runTask(ctx, task)
		step.TaskResult = &task

		if err != nil {
			step.State.Event(ctx, "error")
			step.Err = err

			return step, err
		}
	}

	if step.Tmpl.Kind == domain.StepKindProcess {
		proc := domain.NewProcess(step.Tmpl.ProcessTemplate)
		proc, err := s.procRunner.Start(ctx, proc)
		step.ProcessResult = &proc

		if err != nil {
			step.State.Event(ctx, "error")
			step.Err = err

			return step, err
		}
	}

	step.State.Event(ctx, "completed")
	slog.Debug("Step completed", "step_id", step.ID)

	return step, nil
}

func (s *StepRunner) runTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	nodeID := s.nodeIDRepo.NodeUUID()

	subject := fmt.Sprintf(domain.EventTaskStart, nodeID)
	payload, err := json.Marshal(task)
	if err != nil {
		task.State.Event(ctx, "error")
		task.Err = err

		return task, err
	}

	response, err := s.eventBus.Request(ctx, subject, payload)
	if err != nil {
		task.State.Event(ctx, "error")
		task.Err = err

		return task, err
	}

	var t domain.Task
	if err := json.Unmarshal(response.Data, &task); err != nil {
		task.State.Event(ctx, "error")
		task.Err = err

		return task, err
	}
	task = t

	slog.Debug("Task finished", "task_id", task.ID)
	return task, nil
}
