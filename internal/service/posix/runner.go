package posix

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"remoteMake/internal/model"
	"remoteMake/internal/service"
)

type posixRunner struct {
	uid *service.UIDService
}

func NewRunner(uid *service.UIDService) model.Runner {
	return &posixRunner{uid: uid}
}

func (r *posixRunner) Create(spec model.Spec) (model.Process, error) {
	if len(spec.Command) == 0 {
		return nil, errors.New("no command")
	}

	cmd := exec.Command(spec.Command[0], spec.Command[1:]...)

	cmd.Env = os.Environ()
	for k, v := range spec.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if spec.Dir != "" {
		cmd.Dir = spec.Dir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	proc := &posixProcess{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		done:   make(chan model.State, 1),
	}
	return proc, nil
}
