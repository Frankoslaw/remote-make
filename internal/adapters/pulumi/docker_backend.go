package pulumi

import (
	"context"
	"fmt"
	"log/slog"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DockerNodeManager struct {
	nodeID   uuid.UUID
	eventBus ports.EventBus
}

func NewDockerNodeManager(ni uuid.UUID, ev ports.EventBus) *DockerNodeManager {
	return &DockerNodeManager{nodeID: ni, eventBus: ev}
}

func (n *DockerNodeManager) Provision(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	slog.Debug("Provisioning worker", "worker_id", worker.ID)
	worker.State.Event(ctx, "provision")

	stackName := fmt.Sprintf("remote-make-worker-%s", worker.ID)
	deployFunc := func(ctx *pulumi.Context) error {
		image, err := docker.NewRemoteImage(ctx, fmt.Sprintf("remote-make-image-%s", worker.Tmpl.DockerImage), &docker.RemoteImageArgs{
			Name: pulumi.String(worker.Tmpl.DockerImage),
		})
		if err != nil {
			return err
		}

		container, err := docker.NewContainer(ctx, fmt.Sprintf("remote-make-worker-%s", worker.ID), &docker.ContainerArgs{
			Image: image.ImageId,
			Name:  pulumi.String(fmt.Sprintf("remote-make-worker-%s", worker.ID)),
			Envs: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("NODE_UUID=%s", worker.NodeID)),
			},
			Command: pulumi.StringArray{pulumi.String("sleep"), pulumi.String("infinity")},
		})
		if err != nil {
			return err
		}

		ctx.Export("containerName", container.Name)
		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, "remote-make", deployFunc)
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}

	err = stack.SetConfig(ctx, "docker:host", auto.ConfigValue{Value: "unix:///var/run/docker.sock"})
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}

	_, err = stack.Refresh(ctx)
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}

	res, err := stack.Up(ctx, optup.Message(fmt.Sprintf("Provisioning worker %s", worker.ID)))
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}
	containerName := res.Outputs["containerName"].Value.(string)

	worker.State.Event(ctx, "provisioned")
	slog.Debug("Provisioned remote Docker worker", "worker_id", worker.ID, "container_name", containerName)

	return worker, nil
}

func (n *DockerNodeManager) Terminate(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	slog.Debug("Terminating worker", "worker_id", worker.ID)
	worker.State.Event(ctx, "terminate")

	stackName := fmt.Sprintf("remote-make-worker-%s", worker.ID)
	stack, err := auto.SelectStackInlineSource(ctx, stackName, "remote-make", nil)
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}

	_, err = stack.Destroy(ctx, optdestroy.Message("Destroying worker container"))
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err

		return worker, err
	}

	worker.State.Event(ctx, "terminated")
	slog.Debug("Terminated remote worker", "worker_id", worker.ID, "node_id", worker.NodeID)

	return worker, nil
}
