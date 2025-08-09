//go:build master
// +build master

package pulumi

import (
	"context"
	"fmt"
	"log/slog"
	"remote-make/internal/adapters/config"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"time"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DockerNodeManager struct {
	cfg      *config.Config
	nodeID   uuid.UUID
	eventBus ports.EventBus
}

func NewDockerNodeManager(cfg *config.Config, ni uuid.UUID, ev ports.EventBus) *DockerNodeManager {
	return &DockerNodeManager{cfg: cfg, nodeID: ni, eventBus: ev}
}

func (n *DockerNodeManager) Provision(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	slog.Debug("Provisioning worker", "worker_id", worker.ID)

	worker.State.Event(ctx, "provision")
	worker.NodeID = uuid.New()

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
			Mounts: docker.ContainerMountArray{
				docker.ContainerMountArgs{
					Type:   pulumi.String("bind"),
					Source: pulumi.String(n.cfg.WorkerBinPath),
					Target: pulumi.String("/usr/local/bin/worker"),
				},
			},
			Hosts: docker.ContainerHostArray{
				&docker.ContainerHostArgs{
					Host: pulumi.String("host.docker.internal"),
					Ip:   pulumi.String("host-gateway"),
				},
			},
			Envs: pulumi.StringArray{
				pulumi.String("STEP_RUNNER_ENABLED=true"),
				pulumi.String("EMBEDDED_NATS_ENABLED=false"),
				pulumi.String("NATS_URL=nats://host.docker.internal:4222"),
				pulumi.String(fmt.Sprintf("NODE_UUID=%s", worker.NodeID)),
			},
			Command: pulumi.StringArray{
				pulumi.String("/usr/local/bin/worker"),
			},
			Restart: pulumi.String("always"),
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
		slog.Error(err.Error())

		return worker, err
	}

	err = stack.SetConfig(ctx, "docker:host", auto.ConfigValue{Value: n.cfg.DockerHost})
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err
		slog.Error(err.Error())

		return worker, err
	}
	// err = stack.SetConfig(ctx, "remote-make:workerBinPath", auto.ConfigValue{Value: n.cfg.WorkerBinPath})
	// if err != nil {
	// 	worker.State.Event(ctx, "error")
	// 	worker.Err = err
	// 	return worker, err
	// }

	_, err = stack.Refresh(ctx)
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err
		slog.Error(err.Error())

		return worker, err
	}

	res, err := stack.Up(ctx, optup.Message(fmt.Sprintf("Provisioning worker %s", worker.ID)))
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err
		slog.Error(err.Error())

		return worker, err
	}
	containerName := res.Outputs["containerName"].Value.(string)
	slog.Debug("Provisioned remote Docker worker", "worker_id", worker.ID, "container_name", containerName)

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

healthCheckLoop:
	for {
		select {
		case <-timeoutCtx.Done():
			worker.State.Event(ctx, "error")
			worker.Err = err

			return worker, err
		case <-ticker.C:
			_, err := n.eventBus.Request(ctx, fmt.Sprintf(domain.EventNodeReady, worker.NodeID), []byte("OK"))

			if err != nil {
				slog.Debug("Worker not ready yet", "worker_id", worker.ID, "error", err)
				continue
			} else {
				break healthCheckLoop
			}
		}
	}

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
		slog.Error(err.Error())

		return worker, err
	}

	_, err = stack.Destroy(ctx, optdestroy.Message("Destroying worker container"))
	if err != nil {
		worker.State.Event(ctx, "error")
		worker.Err = err
		slog.Error(err.Error())

		return worker, err
	}

	worker.State.Event(ctx, "terminated")
	slog.Debug("Terminated remote worker", "worker_id", worker.ID, "node_id", worker.NodeID)

	return worker, nil
}
