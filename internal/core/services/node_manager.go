package services

import (
	"context"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
)

type MultiNodeManager struct {
	backends map[string]ports.NodeManager
}

func NewMultiNodeManager() *MultiNodeManager {
	return &MultiNodeManager{backends: make(map[string]ports.NodeManager)}
}

func (n *MultiNodeManager) RegisterBackend(backend string, manager ports.NodeManager) {
	if _, exists := n.backends[backend]; exists {
		panic(fmt.Sprintf("backend %s already registered", backend))
	}
	n.backends[backend] = manager
}

func (n *MultiNodeManager) Provision(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	if n.backends[worker.Tmpl.Backend] == nil {
		worker.State.Event(ctx, "error")
		worker.Err = fmt.Errorf("unknown backend: %s", worker.Tmpl.Backend)

		return worker, worker.Err
	}

	return n.backends[worker.Tmpl.Backend].Provision(ctx, worker)
}

func (n *MultiNodeManager) Terminate(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	if n.backends[worker.Tmpl.Backend] == nil {
		worker.State.Event(ctx, "error")
		worker.Err = fmt.Errorf("unknown backend: %s", worker.Tmpl.Backend)

		return worker, worker.Err
	}

	return n.backends[worker.Tmpl.Backend].Terminate(ctx, worker)
}
