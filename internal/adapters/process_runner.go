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

func (r *LocalProcessRunner) Run(pt domain.ProcessTemplate) (domain.ProcessResult, error) {
	cmd := exec.Command("sh", "-c", pt.Cmd)
	cmd.Dir = pt.Pwd
	if pt.Stdin != "" {
		cmd.Stdin = bytes.NewBufferString(pt.Stdin)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return domain.ProcessResult{
		ID:       uuid.New(),
		ExitCode: exitCode,
		Stdout:   out.String(),
		Stderr:   "", // optional: split stderr from stdout
	}, err
}
