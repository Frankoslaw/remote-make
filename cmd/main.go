package main

import (
	"context"
	"encoding/json"
	"fmt"
	"remote-make/internal/adapters"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/services"
	"time"

	"github.com/google/uuid"
)

func main() {
	// Setup sample templates
	tt := domain.TaskTemplate{
		ID:   uuid.New(),
		Name: "sample-task",
		WorkerTemplate: domain.WorkerTemplate{
			ID:      uuid.New(),
			Name:    "local",
			IsLocal: true,
		},
		StepTemplates: []domain.StepTemplate{
			{
				ID:       uuid.New(),
				SeqOrder: 1,
				TaskTemplate: domain.TaskTemplate{
					ID:   uuid.New(),
					Name: "sample-nested-task",
					WorkerTemplate: domain.WorkerTemplate{
						ID:      uuid.New(),
						Name:    "local",
						IsLocal: true,
					},
					StepTemplates: []domain.StepTemplate{
						{
							ID:       uuid.New(),
							SeqOrder: 1,
							ProcessTemplate: domain.ProcessTemplate{
								ID:  uuid.New(),
								Cmd: "echo 'Hello from sub step 1'",
							},
						},
						{
							ID:       uuid.New(),
							SeqOrder: 2,
							ProcessTemplate: domain.ProcessTemplate{
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
				ProcessTemplate: domain.ProcessTemplate{
					ID:  uuid.New(),
					Cmd: "echo 'Hello from last step 2'",
				},
			},
		},
	}

	// Generic services
	identRepo := adapters.NewNodeIdentityRepo()
	eventBus, err := adapters.NewEmbeddedNatsEventBus()
	if err != nil {
		panic(err)
	}

	// Worker services
	procRunner := adapters.NewLocalProcessRunner()
	stepRunner := services.NewStepRunner(identRepo, eventBus, procRunner)
	stepRunnerHandler := adapters.NewStepRunnerHandler(identRepo, stepRunner)
	stepRunnerHandler.RegisterSubs(eventBus)

	// Master services
	nodeManager := services.NewNodeManager(identRepo, eventBus)
	nodeManagerHandler := adapters.NewNodeManagerHandler(identRepo, nodeManager)
	nodeManagerHandler.RegisterSubs(eventBus)
	taskRunner := services.NewTaskRunner(identRepo, eventBus)

	ctx, cancle := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancle()

	t, err := taskRunner.Start(ctx, tt)

	data, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(data))

	if err != nil {
		fmt.Println("Error:", err)
	}
}
