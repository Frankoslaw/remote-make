package provision

import (
	"context"
	"errors"
	"fmt"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"sync"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DockerProvider struct {
	mu     sync.Mutex
	stacks map[string]auto.Stack
}

func NewDockerProvider() *DockerProvider {
	return &DockerProvider{
		stacks: make(map[string]auto.Stack),
	}
}

func (p *DockerProvider) ID() domain.ProviderID {
	return domain.ProviderID("docker-pulumi")
}

func (p *DockerProvider) Provision(spec domain.NodeSpec) (domain.Node, error) {
	ctx := context.Background()

	image := "ubuntu:24.04"
	if v, ok := spec.Extra["image"]; ok {
		image = v.(string)
	}

	name := uuid.New().String()
	if v, ok := spec.Extra["name"]; ok && v != "" {
		name = v.(string)
	}

	var cpus *string = nil
	if v, ok := spec.Extra["cpus"]; ok {
		cpusStr := fmt.Sprintf("%d", v.(int))
		cpus = &cpusStr
	}
	// TODO: cpu limiting is currently broken on pulumi docker provider due to problem with underlying terraform implementation
	cpus = nil

	var mem *int = nil
	if v, ok := spec.Extra["memory_mb"]; ok {
		memVal := v.(int)
		mem = &memVal
	}

	nodeUUID := uuid.New().String()
	nodeID := domain.NodeID(nodeUUID)
	stackName := fmt.Sprintf("%s-node-%s", p.ID(), nodeUUID)

	pulumiProgram := func(ctx *pulumi.Context) error {
		img, err := docker.NewRemoteImage(ctx, "image", &docker.RemoteImageArgs{
			Name: pulumi.String(image),
		})
		if err != nil {
			return err
		}

		containerArgs := &docker.ContainerArgs{
			Image:   img.ImageId,
			Name:    pulumi.String(name),
			Cpus:    pulumi.StringPtrFromPtr(cpus),
			Memory:  pulumi.IntPtrFromPtr(mem),
			Command: pulumi.ToStringArray([]string{"sleep", "10"}),
		}

		_, err = docker.NewContainer(ctx, name, containerArgs)
		if err != nil {
			return err
		}

		return nil
	}

	stack, err := auto.NewStackInlineSource(ctx, stackName, "remote-make", pulumiProgram)
	if err != nil {
		return domain.Node{}, fmt.Errorf("creating pulumi stack: %w", err)
	}

	if err := stack.Workspace().InstallPlugin(ctx, "docker", "v3.0.0"); err != nil {
		return domain.Node{}, fmt.Errorf("installing docker pulumi plugin: %w", err)
	}

	_, err = stack.Up(ctx)
	if err != nil {
		return domain.Node{}, fmt.Errorf("pulumi up failed: %w", err)
	}

	p.mu.Lock()
	p.stacks[string(nodeID)] = stack
	p.mu.Unlock()

	node := domain.Node{
		NodeID:     nodeID,
		ProviderID: p.ID(),
		Status:     domain.NodeStatusRunning,
		Addr:       "",
		Meta: map[string]any{
			"pulumiStack":     stackName,
			"pulumiStackName": stackName,
		},
		Cap: make(map[domain.Cap]bool),
	}

	return node, nil
}

func (p *DockerProvider) Destroy(nodeID domain.NodeID) error {
	p.mu.Lock()
	stack, ok := p.stacks[string(nodeID)]
	p.mu.Unlock()

	if !ok {
		return errors.New("stack for node not found")
	}

	ctx := context.Background()
	if _, err := stack.Destroy(ctx); err != nil {
		return fmt.Errorf("pulumi destroy failed: %w", err)
	}

	if err := stack.Workspace().RemoveStack(ctx, stack.Name()); err != nil {
		// best-effort; ignore error
	}

	p.mu.Lock()
	delete(p.stacks, string(nodeID))
	p.mu.Unlock()

	return nil
}

func (p *DockerProvider) Controller(node *domain.Node) (port.NodeController, error) {
	return nil, errors.New(
		"docker provider does not support node lifecycle operations",
	)
}
