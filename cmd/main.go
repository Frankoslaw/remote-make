package main

import (
	"fmt"
	"remote-make/internal/adapters"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/services"

	"github.com/google/uuid"
)

func main() {
	// Setup sample templates
	taskT := domain.TaskTemplate{
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
				Type:     domain.NestedTaskStep,
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
							Type:     domain.ProcessStep,
							SeqOrder: 1,
							ProcessTemplate: domain.ProcessTemplate{
								ID:  uuid.New(),
								Cmd: "echo 'Hello from sub step 1'",
							},
						},
						{
							ID:       uuid.New(),
							Type:     domain.ProcessStep,
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
				Type:     domain.ProcessStep,
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
	_ = stepRunner

	// Master services
	nodeManager := services.NewNodeManager(identRepo, eventBus)
	taskRunner := services.NewTaskRunner(identRepo, eventBus, nodeManager)

	task, err := taskRunner.Start(taskT)
	_ = task

	if err != nil {
		fmt.Println("Error:", err)
	}
}
