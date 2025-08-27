package service

import (
	"bytes"
	"context"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"os/exec"
)

func DefaultExec(ctx context.Context, r port.Runner, cmd domain.Command) (domain.ExecResult, error) {
	var outBuf, errBuf bytes.Buffer

	wait, err := r.Attach(ctx, cmd.Args, bytes.NewReader([]byte(cmd.Stdin)), &outBuf, &errBuf)
	if err != nil {
		return domain.ExecResult{ExitCode: -1, Stdout: outBuf.Bytes(), Stderr: errBuf.Bytes(), Err: err.Error()}, err
	}

	err = wait()
	Err := ""
	exitCode := 0

	if err != nil {
		Err = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return domain.ExecResult{ExitCode: exitCode, Stdout: outBuf.Bytes(), Stderr: errBuf.Bytes(), Err: Err}, err
}
