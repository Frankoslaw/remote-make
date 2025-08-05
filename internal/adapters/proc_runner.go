package adapters

import (
	"bytes"
	"os/exec"
	"remote-make/internal/core/domain"

	"github.com/google/uuid"
)

type LocalProcessRunner struct{}

func NewLocalProcessRunner() *LocalProcessRunner {
	return &LocalProcessRunner{}
}

func (r *LocalProcessRunner) Start(pt domain.ProcessTemplate) (domain.Process, error) {
	cmd := exec.Command("sh", "-c", pt.Cmd)
	cmd.Dir = pt.Pwd
	if pt.Stdin != "" {
		cmd.Stdin = bytes.NewBufferString(pt.Stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return domain.Process{
		ID:       uuid.New(),
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, err
}
