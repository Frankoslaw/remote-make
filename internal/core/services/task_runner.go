//go:build master
// +build master

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type TaskRunner struct {
	nodeID   uuid.UUID
	eventBus ports.EventBus
}

func NewTaskRunner(ni uuid.UUID, ev ports.EventBus) *TaskRunner {
	return &TaskRunner{nodeID: ni, eventBus: ev}
}

func (t *TaskRunner) Start(ctx context.Context, task domain.Task) (domain.Task, error) {
	slog.Debug("Starting task", "task_id", task.ID)
	task.State.Event(ctx, "start")

	subject := fmt.Sprintf(domain.EventNodeProvision, t.nodeID)

	worker := domain.NewWorker(&task.Tmpl.WorkerTemplate)
	payload, err := json.Marshal(worker)
	if err != nil {
		task.State.Event(ctx, "error")
		task.Err = err
		slog.Error(err.Error())

		return task, err
	}

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		task.State.Event(ctx, "error")
		task.Err = err
		slog.Error(err.Error())

		return task, err
	}

	var w domain.Worker
	if err := json.Unmarshal(response.Data, &w); err != nil {
		task.State.Event(ctx, "error")
		task.Err = err
		slog.Error(err.Error())

		return task, err
	}
	task.Worker = w

	for _, st := range task.Tmpl.StepTemplates {
		step := domain.NewStep(&st)
		step, err := t.runStep(ctx, task.Worker, step)
		task.Steps = append(task.Steps, step)

		if err != nil {
			task.State.Event(ctx, "error")
			task.Err = err
			slog.Error(err.Error())

			return t.cleanup(ctx, task)
		}
	}

	task.State.Event(ctx, "complete")
	slog.Debug("Task completed", "task_id", task.ID)

	return t.cleanup(ctx, task)
}

func (t *TaskRunner) cleanup(ctx context.Context, task domain.Task) (domain.Task, error) {
	slog.Debug("Cleaning up task", "task_id", task.ID)
	nodeID := t.nodeID

	subject := fmt.Sprintf(domain.EventNodeTerminate, nodeID)
	payload, _ := json.Marshal(task.Worker)

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		task.State.Event(ctx, "error")
		task.Err = err
		slog.Error(err.Error())

		return task, err
	}

	var worker domain.Worker
	if err := json.Unmarshal(response.Data, &worker); err != nil {
		task.State.Event(ctx, "error")
		task.Err = err
		slog.Error(err.Error())

		return task, err
	}

	task.Worker = worker
	slog.Debug("Task cleanup complete", "task_id", task.ID)

	return task, nil
}

func (t *TaskRunner) runStep(ctx context.Context, worker domain.Worker, step domain.Step) (domain.Step, error) {
	nodeID := t.nodeID
	if step.Tmpl.Kind == domain.StepKindProcess {
		nodeID = worker.NodeID
	}

	subject := fmt.Sprintf(domain.EventStepStart, nodeID)
	payload, _ := json.Marshal(step)

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		step.State.Event(ctx, "error")
		step.Err = err
		slog.Error(err.Error())

		return step, err
	}

	var s domain.Step
	if err := json.Unmarshal(response.Data, &s); err != nil {
		step.State.Event(ctx, "error")
		step.Err = err

		return step, err
	}
	step = s

	return step, nil
}
