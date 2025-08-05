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
	wt := domain.WorkerTemplate{ID: uuid.New(), Name: "local", IsLocal: true}
	pt1 := domain.ProcessTemplate{ID: uuid.New(), Cmd: "echo Hello", Pwd: ".", Stdin: ""}
	step1 := domain.StepTemplate{ID: uuid.New(), SeqOrder: 1, ProcessTemplate: pt1}
	taskT := domain.TaskTemplate{ID: uuid.New(), Name: "sample-task", WorkerTemplate: wt, StepTemplates: []domain.StepTemplate{step1}}

	tmplRepo := adapters.NewInMemoryTemplateRepo([]domain.TaskTemplate{taskT})
	runner := adapters.NewLocalProcessRunner()
	tr := services.NewTaskRunner(tmplRepo, runner)
	if err := tr.RunTask(taskT.ID); err != nil {
		fmt.Println("Error:", err)
	}
}
