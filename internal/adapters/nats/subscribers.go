package nats

import (
	"context"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"time"

	"github.com/nats-io/nats.go"
)

type NodeManagerSubscriber struct {
	nodeIDRepo  ports.NodeIdentityRepo
	nodeManager ports.NodeManager
}

func NewNodeManagerSubscriber(ni ports.NodeIdentityRepo, nm ports.NodeManager) *NodeManagerSubscriber {
	return &NodeManagerSubscriber{
		nodeIDRepo:  ni,
		nodeManager: nm,
	}
}

func (s *NodeManagerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventNodeProvision, s.nodeIDRepo.NodeUUID()), s.NodeProvision)
	ev.Subscribe(fmt.Sprintf(domain.EventNodeTerminate, s.nodeIDRepo.NodeUUID()), s.NodeTerminate)
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
	nodeIDRepo ports.NodeIdentityRepo
	taskRunner ports.TaskRunner
}

func NewTaskRunnerSubscriber(ni ports.NodeIdentityRepo, tr ports.TaskRunner) *TaskRunnerSubscriber {
	return &TaskRunnerSubscriber{
		nodeIDRepo: ni,
		taskRunner: tr,
	}
}

func (s *TaskRunnerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventTaskStart, s.nodeIDRepo.NodeUUID()), s.TaskStart)
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
	nodeIDRepo ports.NodeIdentityRepo
	stepRunner ports.StepRunner
}

func NewStepRunnerSubscriber(ni ports.NodeIdentityRepo, sr ports.StepRunner) *StepRunnerSubscriber {
	return &StepRunnerSubscriber{
		nodeIDRepo: ni,
		stepRunner: sr,
	}
}

func (s *StepRunnerSubscriber) RegisterSubscribers(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventStepStart, s.nodeIDRepo.NodeUUID()), s.StepStart)
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
