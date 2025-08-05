package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type TaskRunner struct {
	nodeIDRepo  ports.NodeIdentityRepo
	eventBus    ports.EventBus
	nodeManager ports.NodeManager
}

func NewTaskRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, nm ports.NodeManager) *TaskRunner {
	return &TaskRunner{nodeIDRepo: ni, eventBus: ev, nodeManager: nm}
}

func (t *TaskRunner) Start(ctx context.Context, tt domain.TaskTemplate) (domain.Task, error) {
	// Schedule task
	task := domain.Task{ID: uuid.New(), State: domain.TaskScheduled}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Provision task
	worker, err := t.nodeManager.Provision(ctx, tt.WorkerTemplate)
	if err != nil {
		task.State = domain.TaskError
		return task, err
	}
	task.Worker = worker

	// Run task
	task.State = domain.TaskRunning
	for _, st := range tt.StepTemplates {
		step, err := t.executeStep(ctx, task.Worker, st)
		if err != nil {
			task.State = domain.TaskError
			return t.cleanup(ctx, task)
		}
		if step.State == domain.StepError {
			task.State = domain.TaskError
			return t.cleanup(ctx, task)
		}
		task.Steps = append(task.Steps, step)
	}

	// Task done
	task.State = domain.TaskDone
	return t.cleanup(ctx, task)
}
func (t *TaskRunner) cleanup(ctx context.Context, task domain.Task) (domain.Task, error) {
	w := task.Worker
	if w.State == domain.WorkerProvisioned {
		err := t.nodeManager.Terminate(ctx, w)
		if err != nil {
			task.State = domain.TaskError
			return task, err
		}
	}
	return task, nil
}

func (t *TaskRunner) executeStep(ctx context.Context, w domain.Worker, st domain.StepTemplate) (domain.Step, error) {
	switch st.Type {
	case domain.ProcessStep:
		subject := fmt.Sprintf(domain.EventStepStart, w.NodeID, st.ID)
		payload, _ := json.Marshal(st)

		response, err := t.eventBus.Request(ctx, subject, payload)
		if err != nil {
			return domain.Step{ID: st.ID, State: domain.StepError}, err
		}

		var step domain.Step
		if err := json.Unmarshal(response.Data, &step); err != nil {
			return domain.Step{ID: st.ID, State: domain.StepError}, err
		}

		return step, nil
	case domain.NestedTaskStep:
		task, err := t.Start(ctx, st.TaskTemplate)
		step := domain.Step{ID: uuid.New(), State: domain.StepDone, Task: task}
		if err != nil {
			step.State = domain.StepError
		}
		return step, err
	default:
		return domain.Step{ID: uuid.New(), State: domain.StepError}, errors.New("unsupported step type")
	}
}
