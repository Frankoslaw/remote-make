package nats

import (
	"context"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type NodeManagerSubscriber struct {
	nodeID      uuid.UUID
	nodeManager ports.NodeManager
}

func NewNodeManagerSubscriber(ni uuid.UUID, nm ports.NodeManager) *NodeManagerSubscriber {
	return &NodeManagerSubscriber{
		nodeID:      ni,
		nodeManager: nm,
	}
}

func (s *NodeManagerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventNodeProvision, s.nodeID), s.NodeProvision)
	ev.Subscribe(fmt.Sprintf(domain.EventNodeTerminate, s.nodeID), s.NodeTerminate)
}

func (s *NodeManagerSubscriber) NodeProvision(msg *nats.Msg) {
	go func() {
		// TODO: Hardcoded time
		ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
		defer cancel()

		worker := MustUnmarshal[domain.Worker](msg.Data)
		worker, _ = s.nodeManager.Provision(ctx, worker)

		_ = msg.Respond(MustMarshal(worker))
	}()
}

func (s *NodeManagerSubscriber) NodeTerminate(msg *nats.Msg) {
	go func() {
		// TODO: Hardcoded time
		ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
		defer cancel()

		worker := MustUnmarshal[domain.Worker](msg.Data)
		worker, _ = s.nodeManager.Terminate(ctx, worker)

		_ = msg.Respond(MustMarshal(worker))
	}()
}

type TaskRunnerSubscriber struct {
	nodeID     uuid.UUID
	taskRunner ports.TaskRunner
}

func NewTaskRunnerSubscriber(ni uuid.UUID, tr ports.TaskRunner) *TaskRunnerSubscriber {
	return &TaskRunnerSubscriber{
		nodeID:     ni,
		taskRunner: tr,
	}
}

func (s *TaskRunnerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventTaskStart, s.nodeID), s.TaskStart)
}

func (s *TaskRunnerSubscriber) TaskStart(msg *nats.Msg) {
	go func() {
		// TODO: Hardcoded time
		ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
		defer cancel()

		task := MustUnmarshal[domain.Task](msg.Data)
		task, _ = s.taskRunner.Start(ctx, task)

		_ = msg.Respond(MustMarshal(task))
	}()
}

type StepRunnerSubscriber struct {
	nodeID     uuid.UUID
	stepRunner ports.StepRunner
}

func NewStepRunnerSubscriber(ni uuid.UUID, sr ports.StepRunner) *StepRunnerSubscriber {
	return &StepRunnerSubscriber{
		nodeID:     ni,
		stepRunner: sr,
	}
}

func (s *StepRunnerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventStepStart, s.nodeID), s.StepStart)
}

func (s *StepRunnerSubscriber) StepStart(msg *nats.Msg) {
	go func() {
		// TODO: Hardcoded time
		ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
		defer cancel()

		step := MustUnmarshal[domain.Step](msg.Data)
		step, _ = s.stepRunner.Start(ctx, step)

		_ = msg.Respond(MustMarshal(step))
	}()

}
