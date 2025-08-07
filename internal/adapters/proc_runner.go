package adapters

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"remote-make/internal/core/domain"
)

type LocalProcessRunner struct{}

func NewLocalProcessRunner() *LocalProcessRunner {
	return &LocalProcessRunner{}
}

func (r *LocalProcessRunner) Start(ctx context.Context, process domain.Process) (domain.Process, error) {
	slog.Debug("Starting local process", "process_id", process.ID)
	process.State.Event(ctx, "start")

	cmd := exec.CommandContext(ctx, "sh", "-c", process.Tmpl.Cmd)
	cmd.Dir = process.Tmpl.Pwd
	if process.Tmpl.Stdin != "" {
		cmd.Stdin = bytes.NewBufferString(process.Tmpl.Stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		process.State.Event(ctx, "error")
		process.Err = err

		return process, err
	}

	process.State.Event(ctx, "completed")
	process.ExitCode = cmd.ProcessState.ExitCode()
	process.Stdout = stdout.String()
	process.Stderr = stderr.String()

	slog.Debug("Local process completed", "process_id", process.ID, "exit_code", process.ExitCode)
	return process, err
}
