package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"remote-make/internal/adapters"
	"remote-make/internal/adapters/config"
	"remote-make/internal/adapters/nats"
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
			IsLocal: true,
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
						IsLocal: true,
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

	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	cfg := config.Load()
	slog.Info("Configuration loaded", "config", cfg)

	// Generic services
	identRepo := adapters.NewNodeIdentityRepo()
	eventBus, err := nats.NewEmbeddedNatsEventBus()
	if err != nil {
		panic(err)
	}

	// Master services
	nodeManager := services.NewNodeManager(identRepo, eventBus)
	nodeManagerSubscriber := nats.NewNodeManagerSubscriber(identRepo, nodeManager)
	nodeManagerSubscriber.RegisterSubscribers(eventBus)

	taskRunner := services.NewTaskRunner(identRepo, eventBus)
	taskRunnerSubscriber := nats.NewTaskRunnerSubscriber(identRepo, taskRunner)
	taskRunnerSubscriber.RegisterSubscribers(eventBus)

	// Worker services
	procRunner := adapters.NewLocalProcessRunner()
	stepRunner := services.NewStepRunner(identRepo, eventBus, procRunner)
	stepRunnerSubscriber := nats.NewStepRunnerSubscriber(identRepo, stepRunner)
	stepRunnerSubscriber.RegisterSubscribers(eventBus)

	ctx, cancle := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancle()

	task := domain.NewTask(&tt)
	_, err = taskRunner.Start(ctx, task)

	if err != nil {
		slog.Error(fmt.Sprintf("Error: %v", err))
	}
}
