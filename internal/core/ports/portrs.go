package ports

import (
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
}

type NodeManager interface {
	Provision(wt domain.WorkerTemplate) (domain.Worker, error)
	Terminate(w domain.Worker) error
}

type TaskRunner interface {
	Start(tt domain.TaskTemplate) (domain.Task, error)
}

type StepRunner interface {
	Start(st domain.StepTemplate) (domain.Step, error)
}

type ProcessRunner interface {
	Start(pt domain.ProcessTemplate) (domain.Process, error)
}
