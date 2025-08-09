//go:build master
// +build master

package pulumi

import (
	"context"
	"fmt"
	"log/slog"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type LocalNodeManager struct {
	nodeID   uuid.UUID
	eventBus ports.EventBus
}

func NewLocalNodeManager(ni uuid.UUID, ev ports.EventBus) *LocalNodeManager {
	return &LocalNodeManager{nodeID: ni, eventBus: ev}
}

func (n *LocalNodeManager) Provision(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	slog.Debug("Provisioning worker", "worker_id", worker.ID)
	worker.State.Event(ctx, "provision")

	worker.State.Event(ctx, "provisioned")
	worker.NodeID = n.nodeID

	slog.Debug("Provisioned local worker", "worker_id", worker.ID, "node_id", worker.NodeID)
	return worker, nil
}

func (n *LocalNodeManager) Terminate(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	slog.Debug("Terminating worker", "worker_id", worker.ID)
	worker.State.Event(ctx, "terminate")

	if worker.NodeID != n.nodeID {
		worker.State.Event(ctx, "error")
		worker.Err = fmt.Errorf("worker %s is not managed by this node", worker.ID)
		slog.Error(worker.Err.Error())

		return worker, worker.Err
	}

	worker.State.Event(ctx, "terminated")
	slog.Debug("Terminated local worker", "worker_id", worker.ID, "node_id", worker.NodeID)

	return worker, nil
}
