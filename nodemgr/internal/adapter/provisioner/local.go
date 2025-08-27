package provisioner

import (
	"context"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"os"
)

type LocalNode struct {
	domain.Node
}

func (n *LocalNode) Conn() (*domain.NodeConn, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	n.Node.Conn.Hostname = hostname
	return &n.Node.Conn, nil
}

var _ port.NodeClient = (*LocalNode)(nil)

type LocalProvisioner struct{}

func NewLocalProvisioner() *LocalProvisioner {
	return &LocalProvisioner{}
}

func (l *LocalProvisioner) ID() string {
	return "local"
}

func (l *LocalProvisioner) Up(ctx context.Context, spec map[string]string) (port.NodeClient, error) {
	return &LocalNode{
		Node: domain.Node{
			NodeID:   "local",
			Status:   domain.NodeStatusRunning,
			Provider: l.ID(),
			Conn: domain.NodeConn{
				PublicIP:  "localhost",
				PrivateIP: "localhost",
			},
			Caps: map[string]bool{"local.enabled": true},
			Tags: make(map[string]string),
		},
	}, nil
}

func (l *LocalProvisioner) Destroy(ctx context.Context, nodeID string) error {
	return nil
}

var _ port.Provisioner = (*LocalProvisioner)(nil)
