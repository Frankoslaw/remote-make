package adapters

import "nodemgr/internal/core/domain"

type InfraBackend interface {
	Create()
	Destroy()
	Conn() domain.NodeConn

	// optional for providers like docker or local
	Exec()
	Attach()
	Copy()
	Mount()
}

type LocalBackend struct {
}

func (l *LocalBackend) Create()  {}
func (l *LocalBackend) Destroy() {}
func (l *LocalBackend) Conn() domain.NodeConn {
	return domain.NodeConn{
		PublicIP:  "127.0.0.1",
		PrivateIP: "192.168.1.1",
		Hostname:  "localhost",
	}
}
