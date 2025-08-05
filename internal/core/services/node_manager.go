package services

import (
	"errors"
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

func (n *NodeManager) Provision(wt domain.WorkerTemplate) (domain.Worker, error) {
	if wt.IsLocal {
		return domain.Worker{
			ID:     uuid.New(),
			NodeID: n.nodeIDRepo.NodeUUID(),
			State:  domain.WorkerProvisioned,
		}, errors.New("you are trying to provision yourself????")
	}

	panic("unimplemented")
}

func (n *NodeManager) Terminate(w domain.Worker) error {
	if w.NodeID == n.nodeIDRepo.NodeUUID() {
		return nil
	}

	panic("unimplemented")
}
