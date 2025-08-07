package ports

import (
	"context"
	"remote-make/internal/core/domain"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type NodeIdentityRepo interface {
	NodeUUID() uuid.UUID
}

type EventBus interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, handler func(msg *nats.Msg)) error
	Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error)
}

type NodeManager interface {
	Provision(ctx context.Context, worker domain.Worker) (domain.Worker, error)
	Terminate(ctx context.Context, worker domain.Worker) (domain.Worker, error)
}

type TaskRunner interface {
	Start(ctx context.Context, task domain.Task) (domain.Task, error)
}

type StepRunner interface {
	Start(ctx context.Context, step domain.Step) (domain.Step, error)
}

type ProcessRunner interface {
	Start(ctx context.Context, process domain.Process) (domain.Process, error)
}
