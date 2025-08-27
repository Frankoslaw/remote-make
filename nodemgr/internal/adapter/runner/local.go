package runner

import (
	"context"
	"errors"
	"io"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/service"
	"os"
	"os/exec"
)

type LocalRunner struct{}

func NewLocalRunner(node port.NodeClient) (*LocalRunner, error) {
	if !node.HasCap("local.enabled") {
		return nil, errors.New("node does not support local execution")
	}

	return &LocalRunner{}, nil
}

func (l *LocalRunner) ID() string { return "local" }

func (l *LocalRunner) Attach(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer) (func() error, error) {
	if len(cmd) == 0 {
		return nil, exec.ErrNotFound
	}

	proc := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	proc.Stdin = stdin
	proc.Stdout = stdout
	proc.Stderr = stderr

	if err := proc.Start(); err != nil {
		return nil, err
	}

	return func() error { return proc.Wait() }, nil
}

func (l *LocalRunner) Copy(ctx context.Context, src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}

func (l *LocalRunner) Exec(ctx context.Context, cmd domain.Command) (domain.ExecResult, error) {
	return service.DefaultExec(ctx, l, cmd)
}

var _ port.Runner = (*LocalRunner)(nil)
