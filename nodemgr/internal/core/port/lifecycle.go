package port

import "nodemgr/internal/core/domain"

type NodeLifecycle interface {
	OpenLifecycleHandle(node *domain.Node) (NodeLifecycleHandle, error)
}

type NodeLifecycleHandle interface {
	Start()
	Stop() // TODO: Hibernate???
	Reboot()
	Terminate()
}

type NodeLifecycleService interface {
	StartNode(nodeID domain.NodeID) error
	StopNode(nodeID domain.NodeID) error
	RebootNode(nodeID domain.NodeID) error
	TerminateNode(nodeID domain.NodeID) error
}
