package services

import (
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"

	"github.com/google/uuid"
)

type TaskRunner struct {
	tmplRepo ports.TemplateRepo
	runner   ports.ProcessRunner
}

func NewTaskRunner(tr ports.TemplateRepo, pr ports.ProcessRunner) *TaskRunner {
	return &TaskRunner{tmplRepo: tr, runner: pr}
}

func (t *TaskRunner) RunTask(templateID uuid.UUID) error {
	tmpl, err := t.tmplRepo.GetTaskTemplate(templateID)
	if err != nil {
		return err
	}

	// Setup Worker
	nodeID := uuid.New()
	worker := domain.Worker{
		ID:    uuid.New(),
		State: domain.WorkerScheduled,
	}
	if tmpl.WorkerTemplate.IsLocal {
		worker.NodeID = nodeID
		worker.State = domain.WorkerProvisioned
	} else {
		// TODO: Run Provisioner Service
	}
	fmt.Printf("Worker %s provisioned (state=%d)\n", tmpl.WorkerTemplate.Name, worker.State)

	// Create and schedule task
	task := domain.Task{
		ID:     uuid.New(),
		State:  domain.TaskScheduled,
		Worker: worker,
	}
	fmt.Printf("Task %s scheduled (state=%d)\n", tmpl.Name, task.State)

	// Run task
	task.State = domain.TaskRunning
	fmt.Printf("Task %s running (state=%d)\n", tmpl.Name, task.State)

	for _, stepT := range tmpl.StepTemplates {
		// Schedule step
		step := domain.Step{
			ID:    uuid.New(),
			State: domain.StepScheduled,
		}
		fmt.Printf("Step %d scheduled (state=%d): %s\n", stepT.SeqOrder, step.State, stepT.ProcessTemplate.Cmd)

		// Run step
		step.State = domain.StepRunning
		fmt.Printf("Step %d running (state=%d)\n", stepT.SeqOrder, step.State)

		res, err := t.runner.Run(stepT.ProcessTemplate)
		step.ProcessResult = res

		if err != nil || res.ExitCode != 0 {
			step.State = domain.StepError
			task.State = domain.TaskError
			fmt.Printf("Step %d error (state=%d): exit=%d err=%v\n", stepT.SeqOrder, step.State, res.ExitCode, err)

			worker.State = domain.WorkerTerminating
			fmt.Printf("Worker terminating (state=%d)\n", worker.State)

			worker.State = domain.WorkerTerminated
			fmt.Printf("Worker terminated (state=%d)\n", worker.State)

			return fmt.Errorf("task %s failed at step %d\n ", tmpl.Name, stepT.SeqOrder)
		}

		// Step done
		step.State = domain.StepDone
		task.Steps = append(task.Steps, step)
		fmt.Printf("Step %d done (state=%d)\n", stepT.SeqOrder, step.State)

		fmt.Printf("Step %d result:\n	exit_code: %d,\n	stdout: %s\n", stepT.SeqOrder, step.ProcessResult.ExitCode, step.ProcessResult.Stdout)
	}

	// Task done
	task.State = domain.TaskDone
	fmt.Printf("Task %s done (state=%d)\n", tmpl.Name, task.State)

	// Terminate worker
	worker.State = domain.WorkerTerminating
	fmt.Printf("Worker terminating (state=%d)\n", worker.State)

	worker.State = domain.WorkerTerminated
	fmt.Printf("Worker terminated (state=%d)\n", worker.State)

	return nil
}
