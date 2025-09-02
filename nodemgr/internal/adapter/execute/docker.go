package execute

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
)

type DockerExecProvider struct {
	execHandleRepository port.ExecHandleRepository
}

func NewDockerExecProvider(execHandleRepository port.ExecHandleRepository) *DockerExecProvider {
	return &DockerExecProvider{
		execHandleRepository: execHandleRepository,
	}
}

func (p *DockerExecProvider) ID() domain.ExecProviderID {
	return domain.ExecProviderID("docker")
}

func (p *DockerExecProvider) OpenExecHandle(node *domain.Node) (port.ExecHandle, error) {
	if !node.HasCap("exec:docker") {
		return nil, fmt.Errorf("node does not have exec:docker capability")
	}

	containerID := node.Meta["container_id"].(string)
	dockerHost := node.Meta["docker_host"].(string)

	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(dockerHost))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	execHandle := NewDockerExecHandle(client, containerID)
	p.execHandleRepository.Create(execHandle)

	return execHandle, nil
}

var _ port.NodeExecProvider = (*DockerExecProvider)(nil)

type DockerExecHandle struct {
	id          string
	cli         *client.Client
	containerID string
	user        string
}

func NewDockerExecHandle(cli *client.Client, containerID string) *DockerExecHandle {
	ctx := context.Background()

	inspectResp, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil
	}
	user := inspectResp.Config.User

	return &DockerExecHandle{
		id:          uuid.New().String(),
		cli:         cli,
		containerID: containerID,
		user:        user,
	}
}

func (d DockerExecHandle) ID() domain.ExecHandleID {
	return domain.ExecHandleID(d.id)
}

func (d *DockerExecHandle) Close() error {
	if d.cli != nil {
		return d.cli.Close()
	}
	return nil
}

func (d *DockerExecHandle) Attach(req domain.AttachRequest) (*domain.AttachResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *DockerExecHandle) Exec(req domain.ExecRequest) (*domain.ExecResult, error) {
	ctx := context.Background()
	var env []string
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	opts := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          req.Command,
		Env:          env,
		WorkingDir:   req.WorkingDir,
	}

	execID, err := d.cli.ContainerExecCreate(ctx, d.containerID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	var execResult domain.ExecResult
	hijack, err := d.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer hijack.Close()

	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, hijack.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return &execResult, err
		}
		break

	case <-ctx.Done():
		return &execResult, ctx.Err()
	}

	stdout, err := io.ReadAll(&outBuf)
	if err != nil {
		return &execResult, err
	}
	stderr, err := io.ReadAll(&errBuf)
	if err != nil {
		return &execResult, err
	}

	res, err := d.cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return &execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.Stdout = stdout
	execResult.Stderr = stderr

	return &execResult, nil
}

func (d *DockerExecHandle) ExecStream(exec domain.ExecRequest, attach domain.AttachRequest) (*domain.AttachResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *DockerExecHandle) CopyTo(src io.Reader, dst string) error {
	return fmt.Errorf("not implemented")
}

func (d *DockerExecHandle) CopyFrom(src string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

var _ port.ExecHandle = (*DockerExecHandle)(nil)
