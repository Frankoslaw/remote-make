package services

import (
	"context"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type NodeManager struct {
	nodeIDRepo ports.NodeIdentityRepo
	eventBus   ports.EventBus
}

func NewNodeManager(ni ports.NodeIdentityRepo, ev ports.EventBus) *NodeManager {
	return &NodeManager{nodeIDRepo: ni, eventBus: ev}
}

func (n *NodeManager) Provision(ctx context.Context, wt domain.WorkerTemplate) (domain.Worker, error) {
	if wt.IsLocal {
		return domain.Worker{
			ID:     uuid.New(),
			State:  domain.WorkerProvisioned,
			NodeID: n.nodeIDRepo.NodeUUID(),
		}, nil
	}

	panic("unimplemented")
}

func (n *NodeManager) Terminate(ctx context.Context, w domain.Worker) error {
	if w.NodeID == n.nodeIDRepo.NodeUUID() {
		return nil
	}

	panic("unimplemented")
}
