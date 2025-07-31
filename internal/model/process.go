package model

import (
	"context"
	"io"
)

type State struct {
	ExitCode int
	Err      error
}

type Process interface {
	ID() string
	Stdin() io.WriteCloser
	Stdout() io.Reader
	Stderr() io.Reader
	Start(ctx context.Context) (<-chan State, error)
	Wait() (State, error)
	Stop() error
	OnExit(cb func(State))
}

type Runner interface {
	Create(spec Spec) (Process, error)
}

type Spec struct {
	Command []string
	Env     map[string]string
	Dir     string
	Image   string
	Name    string
}
