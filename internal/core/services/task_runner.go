package services

import (
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type TaskRunner struct {
	nodeIDRepo  ports.NodeIdentityRepo
	eventBus    ports.EventBus
	nodeManager ports.NodeManager
}

func NewTaskRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, nm ports.NodeManager) *TaskRunner {
	return &TaskRunner{nodeIDRepo: ni, eventBus: ev, nodeManager: nm}
}

func (t *TaskRunner) Start(tt domain.TaskTemplate) (domain.Task, error) {
	nodeID := t.nodeIDRepo.NodeUUID()

	// Schedule task
	task := domain.Task{
		ID:    uuid.New(),
		State: domain.TaskScheduled,
	}
	fmt.Printf("Task %s scheduled (state=%d)\n", tt.Name, task.State)

	// Provisioning step
	var worker domain.Worker
	if tt.WorkerTemplate.IsLocal {
		worker = domain.Worker{
			ID:     uuid.New(),
			State:  domain.WorkerProvisioned,
			NodeID: nodeID,
		}
	} else {
		w, err := t.nodeManager.Provision(tt.WorkerTemplate)
		if err != nil {
			task.State = domain.TaskError
			return task, err
		}
		worker = w
	}

	task.Worker = worker
	fmt.Printf("Worker %s provisioned (state=%d)\n", tt.WorkerTemplate.Name, worker.State)

	// Run task
	task.State = domain.TaskRunning
	fmt.Printf("Task %s running (state=%d)\n", tt.Name, task.State)

	sort.Slice(tt.StepTemplates, func(i, j int) bool {
		return tt.StepTemplates[i].SeqOrder < tt.StepTemplates[j].SeqOrder
	})

	for _, stepT := range tt.StepTemplates {
		// Schedule step
		step := domain.Step{
			ID:    uuid.New(),
			State: domain.StepScheduled,
		}
		fmt.Printf("Step %d scheduled (state=%d): %s\n", stepT.SeqOrder, step.State, stepT.ProcessTemplate.Cmd)

		doneCh := make(chan struct{})
		var once sync.Once

		// Subscribe to StepDone
		doneSubject := fmt.Sprintf(domain.EventStepDone, step.ID)
		t.eventBus.Subscribe(doneSubject, func(msg *nats.Msg) {
			once.Do(func() {
				json.Unmarshal(msg.Data, &step)
				fmt.Printf("Step %d done (state=%d)\n", stepT.SeqOrder, step.State)
				close(doneCh)
			})
		})

		// Subscribe to StepError
		errorSubject := fmt.Sprintf(domain.EventStepError, step.ID)
		t.eventBus.Subscribe(errorSubject, func(msg *nats.Msg) {
			once.Do(func() {
				json.Unmarshal(msg.Data, &step)
				fmt.Printf("Step %d errored (state=%d)\n", stepT.SeqOrder, step.State)
				close(doneCh)
			})
		})

		// Trigger step start
		step.State = domain.StepRunning
		fmt.Printf("Step %d running (state=%d)\n", stepT.SeqOrder, step.State)

		msgData, _ := json.Marshal(stepT)
		err := t.eventBus.Publish(fmt.Sprintf(domain.EventStepStart, worker.NodeID, step.ID), msgData)
		if err != nil {
			task.State = domain.TaskError
			task, _ = t.cleanUp(task)
			return task, err
		}

		// Wait for either done or error
		select {
		case <-doneCh:
			if step.State == domain.StepError {
				task.State = domain.TaskError
				task, _ = t.cleanUp(task)
				return task, fmt.Errorf("step %d failed", stepT.SeqOrder)
			}
		case <-time.After(30 * time.Second):
			task.State = domain.TaskError
			task, _ = t.cleanUp(task)
			return task, fmt.Errorf("step %d timed out", stepT.SeqOrder)
		}
	}

	// Task done
	task.State = domain.TaskDone
	fmt.Printf("Task %s done (state=%d)\n", tt.Name, task.State)

	task, err := t.cleanUp(task)
	return task, err
}

func (t *TaskRunner) cleanUp(task domain.Task) (domain.Task, error) {
	w := task.Worker

	if w.State == domain.WorkerProvisioned {
		w.State = domain.WorkerTerminating
		fmt.Printf("Worker terminating (state=%d)\n", w.State)

		err := t.nodeManager.Terminate(w)
		if err != nil {
			fmt.Printf("FAILED DURING CLEANUP: %v\n", err)
			task.State = domain.TaskError
			return task, err
		}

		w.State = domain.WorkerTerminated
		fmt.Printf("Worker terminated (state=%d)\n", w.State)

		task.Worker = w
	}

	return task, nil
}
