package services

import (
	"context"
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type StepRunner struct {
	nodeIDRepo ports.NodeIdentityRepo
	eventBus   ports.EventBus
	procRunner ports.ProcessRunner
}

func NewStepRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, pr ports.ProcessRunner) *StepRunner {
	return &StepRunner{nodeIDRepo: ni, eventBus: ev, procRunner: pr}
}

func (s *StepRunner) Start(ctx context.Context, st domain.StepTemplate) (domain.Step, error) {
	step := domain.Step{
		ID:    uuid.New(),
		State: domain.StepScheduled,
	}

	if st.TaskTemplate.ID != uuid.Nil {
		task, err := s.startTask(ctx, st.TaskTemplate)
		step.Task = task

		if err != nil {
			step.State = domain.StepError
			return step, err
		}
		if task.State == domain.TaskError {
			step.State = domain.StepError
			return step, err
		}
	}

	if st.ProcessTemplate.ID != uuid.Nil {
		res, err := s.procRunner.Start(ctx, st.ProcessTemplate)
		step.ProcessResult = res

		if err != nil {
			step.State = domain.StepError
		}
	}

	if step.State != domain.StepError {
		step.State = domain.StepDone
	}

	return step, nil
}

func (s *StepRunner) startTask(ctx context.Context, tt domain.TaskTemplate) (domain.Task, error) {
	nodeID := s.nodeIDRepo.NodeUUID()

	subject := fmt.Sprintf(domain.EventTaskStart, nodeID)
	payload, _ := json.Marshal(tt)

	response, err := s.eventBus.Request(ctx, subject, payload)
	if err != nil {
		return domain.Task{ID: uuid.New(), State: domain.TaskError}, err
	}

	var task domain.Task
	if err := json.Unmarshal(response.Data, &task); err != nil {
		return domain.Task{ID: uuid.New(), State: domain.TaskError}, err
	}

	return task, nil
}
