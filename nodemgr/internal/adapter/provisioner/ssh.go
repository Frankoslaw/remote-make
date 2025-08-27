package provisioner

import (
	"context"
	"strconv"

	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"

	"github.com/google/uuid"
)

type SSHNode struct {
	domain.Node
}

func (n *SSHNode) Conn() (*domain.NodeConn, error) {
	return &n.Node.Conn, nil
}

var _ port.NodeClient = (*SSHNode)(nil)

type SSHProvisioner struct{}

func NewSSHProvisioner() *SSHProvisioner {
	return &SSHProvisioner{}
}

func (p *SSHProvisioner) ID() string {
	return "ssh"
}

func (p *SSHProvisioner) Up(ctx context.Context, spec map[string]string) (port.NodeClient, error) {
	sshUser := spec["ssh_user"]
	if sshUser == "" {
		sshUser = "localuser"
	}

	sshPass := spec["ssh_pass"]
	if sshPass == "" {
		sshPass = "localpass"
	}

	sshPort := 22
	if val, ok := spec["ssh_port"]; ok && val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			sshPort = parsed
		}
	}
	sshKey := spec["ssh_key"]

	return &SSHNode{
		Node: domain.Node{
			NodeID:   uuid.New().String(),
			Status:   domain.NodeStatusRunning,
			Provider: p.ID(),
			Conn: domain.NodeConn{
				PublicIP:  spec["ssh_ip"],
				PrivateIP: spec["ssh_private_ip"],
				SSHUser:   sshUser,
				SSHPass:   sshPass,
				SSHPort:   sshPort,
				SSHKey:    sshKey,
			},
			Caps: map[string]bool{"ssh.enabled": true},
			Tags: make(map[string]string),
		},
	}, nil
}

func (p *SSHProvisioner) Destroy(ctx context.Context, nodeID string) error {
	return nil
}

var _ port.Provisioner = (*SSHProvisioner)(nil)
