package provision

import (
	"context"
	"errors"
	"fmt"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/util"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DockerArgs struct {
	Name      string           `mapstructure:"name"`
	User      string           `mapstructure:"user,omitempty"`
	Image     string           `mapstructure:"image" validate:"required"`
	ImageType domain.ImageType `mapstructure:"image_type" validate:"required"`
	CPUs      int              `mapstructure:"cpus,omitempty"`
	MemoryMB  int              `mapstructure:"memory_mb,omitempty"`
	Command   []string         `mapstructure:"command,omitempty"`
	StdinOpen bool             `mapstructure:"stdin_open,omitempty"`
	Tty       bool             `mapstructure:"tty,omitempty"`
}

type DockerProvider struct {
	mu       sync.Mutex
	stacks   map[string]auto.Stack
	validate *validator.Validate
}

func NewDockerProvider(dockerHost string) *DockerProvider {
	return &DockerProvider{
		stacks:   make(map[string]auto.Stack),
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (p *DockerProvider) ID() domain.ProviderID {
	return domain.ProviderID("docker-pulumi")
}

func (p *DockerProvider) Provision(spec domain.NodeSpec) (*domain.Node, error) {
	ctx := context.Background()

	args, err := util.DecodeExtraTo[DockerArgs](spec.Extra)
	if err != nil {
		return nil, fmt.Errorf("decode extra: %w", err)
	}

	err = p.validate.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("validate args: %w", err)
	}

	if args.Name == "" {
		args.Name = fmt.Sprintf("node-%s", uuid.New().String()[:8])
	}

	if args.Command != nil {
		args.Command = strings.Split(args.Command[0], " ")
	} else {
		args.Command = []string{"sleep", "infinity"}
	}

	nodeUUID := uuid.New().String()
	nodeID := domain.NodeID(nodeUUID)
	stackName := fmt.Sprintf("%s-node-%s", p.ID(), nodeUUID)

	pulumiProgram := func(ctx *pulumi.Context) error {
		img, err := docker.NewRemoteImage(ctx, "image", &docker.RemoteImageArgs{
			Name: pulumi.String(args.Image),
		})
		if err != nil {
			return err
		}

		containerArgs := &docker.ContainerArgs{
			Image: img.ImageId,
			User:  pulumi.StringPtrFromPtr(&args.User),
			Name:  pulumi.String(args.Name),
			// TODO: Cpu flag is currently broken in pulumi bindings as underlying terraform provider does not expose it
			// Cpus:    pulumi.StringPtrFromPtr(cpus),
			Memory:    pulumi.IntPtrFromPtr(&args.MemoryMB),
			Command:   pulumi.ToStringArray(args.Command),
			StdinOpen: pulumi.BoolPtr(args.StdinOpen),
			Tty:       pulumi.BoolPtr(args.Tty),
		}

		container, err := docker.NewContainer(ctx, args.Name, containerArgs)
		if err != nil {
			return err
		}

		ctx.Export("container_id", container.ID())

		return nil
	}

	stack, err := auto.NewStackInlineSource(ctx, stackName, "remote-make", pulumiProgram)
	if err != nil {
		return nil, fmt.Errorf("creating pulumi stack: %w", err)
	}

	if err := stack.Workspace().InstallPlugin(ctx, "docker", "v3.0.0"); err != nil {
		return nil, fmt.Errorf("installing docker pulumi plugin: %w", err)
	}

	upRes, err := stack.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("pulumi up failed: %w", err)
	}

	containerId, ok := upRes.Outputs["container_id"].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to get container_id output from pulumi stack")
	}

	p.mu.Lock()
	p.stacks[string(nodeID)] = stack
	p.mu.Unlock()

	node := domain.Node{
		NodeID:     nodeID,
		ProviderID: p.ID(),
		State:      domain.NodeStateRunning,
		Meta: map[string]any{
			"pulumi_stack":      stackName,
			"pulumi_stack_name": stackName,
			"docker_host":       "unix:///var/run/docker.sock",
			"container_id":      containerId,
		},
		Cap: map[domain.Cap]bool{
			"exec:docker": true,
		},
	}

	return &node, nil
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

var _ port.NodeProvider = (*DockerProvider)(nil)
