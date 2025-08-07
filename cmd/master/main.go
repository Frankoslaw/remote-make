package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"remote-make/internal/adapters"
	"remote-make/internal/adapters/config"
	"remote-make/internal/adapters/nats"
	"remote-make/internal/adapters/pulumi"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/services"
	"time"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
)

func main() {
	// Setup sample templates
	tt := domain.TaskTemplate{
		ID: uuid.New(),
		WorkerTemplate: domain.WorkerTemplate{
			ID:      uuid.New(),
			Backend: "local",
		},
		StepTemplates: []domain.StepTemplate{
			{
				ID:       uuid.New(),
				SeqOrder: 1,
				Kind:     domain.StepKindTask,
				TaskTemplate: &domain.TaskTemplate{
					ID: uuid.New(),
					WorkerTemplate: domain.WorkerTemplate{
						ID:      uuid.New(),
						Backend: "local",
					},
					StepTemplates: []domain.StepTemplate{
						{
							ID:       uuid.New(),
							SeqOrder: 1,
							Kind:     domain.StepKindProcess,
							ProcessTemplate: &domain.ProcessTemplate{
								ID:  uuid.New(),
								Cmd: "echo 'Hello from sub step 1'",
							},
						},
						{
							ID:       uuid.New(),
							SeqOrder: 2,
							Kind:     domain.StepKindProcess,
							ProcessTemplate: &domain.ProcessTemplate{
								ID:  uuid.New(),
								Cmd: "echo 'Hello from sub step 2'",
							},
						},
					},
				},
			},
			{
				ID:       uuid.New(),
				SeqOrder: 2,
				Kind:     domain.StepKindProcess,
				ProcessTemplate: &domain.ProcessTemplate{
					ID:  uuid.New(),
					Cmd: "echo 'Hello from last step 2'",
				},
			},
		},
	}

	// Generic services
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))

	cfg := config.Load()
	slog.Info("Configuration loaded", "config", cfg)

	if cfg.EmbeddedNATSEnabled {
		ns, err := nats.NewEmbeddedNats()
		if err != nil {
			slog.Error("Failed to start embedded NATS server", "error", err)
			return
		}
		defer ns.Shutdown()
		cfg.NATSURL = ns.ClientURL()
		slog.Info("Embedded NATS server started", "url", cfg.NATSURL)
	}

	eventBus, err := nats.NewNatsEventBus(cfg)
	if err != nil {
		slog.Error("Failed to connect to NATS server", "error", err)
		return
	}
	defer eventBus.Shutdown()
	slog.Info("Connected to NATS server", "url", cfg.NATSURL)

	// Master services
	localNodeManager := pulumi.NewLocalNodeManager(cfg.NodeID, eventBus)
	dockerNodeManager := pulumi.NewDockerNodeManager(cfg.NodeID, eventBus)

	nodeManager := services.NewMultiNodeManager()
	nodeManager.RegisterBackend("local", localNodeManager)
	nodeManager.RegisterBackend("docker", dockerNodeManager)

	nodeManagerSubscriber := nats.NewNodeManagerSubscriber(cfg.NodeID, nodeManager)
	nodeManagerSubscriber.RegisterSubscribers(eventBus)

	taskRunner := services.NewTaskRunner(cfg.NodeID, eventBus)
	taskRunnerSubscriber := nats.NewTaskRunnerSubscriber(cfg.NodeID, taskRunner)
	taskRunnerSubscriber.RegisterSubscribers(eventBus)

	// Worker services
	procRunner := adapters.NewLocalProcessRunner()
	stepRunner := services.NewStepRunner(cfg.NodeID, eventBus, procRunner)
	stepRunnerSubscriber := nats.NewStepRunnerSubscriber(cfg.NodeID, stepRunner)
	stepRunnerSubscriber.RegisterSubscribers(eventBus)

	ctx, cancle := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancle()

	task := domain.NewTask(&tt)
	task, err = taskRunner.Start(ctx, task)
	if err != nil {
		slog.Error(fmt.Sprintf("Error: %v", err))
	}

	taskString, _ := json.MarshalIndent(task, "", "\t")
	slog.Info(fmt.Sprintf("Task: %s", string(taskString)))
}
