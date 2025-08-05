package services

import (
	"context"
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type TaskRunner struct {
	nodeIDRepo ports.NodeIdentityRepo
	eventBus   ports.EventBus
}

func NewTaskRunner(ni ports.NodeIdentityRepo, ev ports.EventBus) *TaskRunner {
	return &TaskRunner{nodeIDRepo: ni, eventBus: ev}
}

func (t *TaskRunner) Start(ctx context.Context, tt domain.TaskTemplate) (domain.Task, error) {
	nodeID := t.nodeIDRepo.NodeUUID()

	// Schedule task
	task := domain.Task{ID: uuid.New(), State: domain.TaskScheduled}

	// Provision task
	subject := fmt.Sprintf(domain.EventNodeProvision, nodeID)
	payload, _ := json.Marshal(tt.WorkerTemplate)

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		return domain.Task{ID: task.ID, State: domain.TaskError}, err
	}

	var worker domain.Worker
	if err := json.Unmarshal(response.Data, &worker); err != nil {
		return domain.Task{ID: task.ID, State: domain.TaskError}, err
	}
	task.Worker = worker

	// Run task
	task.State = domain.TaskRunning
	for _, st := range tt.StepTemplates {
		step, err := t.startStep(ctx, task.Worker, st)
		task.Steps = append(task.Steps, step)

		if err != nil {
			task.State = domain.TaskError
			return t.cleanup(ctx, task)
		}
		if step.State == domain.StepError {
			task.State = domain.TaskError
			return t.cleanup(ctx, task)
		}
	}

	// Task done
	task.State = domain.TaskDone
	return t.cleanup(ctx, task)
}
func (t *TaskRunner) cleanup(ctx context.Context, task domain.Task) (domain.Task, error) {
	nodeID := t.nodeIDRepo.NodeUUID()

	subject := fmt.Sprintf(domain.EventNodeTerminate, nodeID)
	payload, _ := json.Marshal(task.Worker)

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		task.State = domain.TaskError
		return task, err
	}

	var worker domain.Worker
	if err := json.Unmarshal(response.Data, &worker); err != nil {
		task.State = domain.TaskError
		return task, err
	}

	task.Worker = worker
	return task, nil
}

func (t *TaskRunner) startStep(ctx context.Context, w domain.Worker, st domain.StepTemplate) (domain.Step, error) {
	nodeID := t.nodeIDRepo.NodeUUID()
	if st.ProcessTemplate.ID != uuid.Nil {
		nodeID = w.NodeID
	}

	subject := fmt.Sprintf(domain.EventStepStart, nodeID)
	payload, _ := json.Marshal(st)

	response, err := t.eventBus.Request(ctx, subject, payload)
	if err != nil {
		return domain.Step{ID: uuid.New(), State: domain.StepError}, err
	}

	var step domain.Step
	if err := json.Unmarshal(response.Data, &step); err != nil {
		return domain.Step{ID: uuid.New(), State: domain.StepError}, err
	}

	return step, nil
}
