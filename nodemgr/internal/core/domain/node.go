package domain

type ProviderID string
type NodeID string
type Cap string

type NodeStatus string

const (
	NodeStatusPending      NodeStatus = "pending"
	NodeStatusRunning      NodeStatus = "running"
	NodeStatusStopping     NodeStatus = "stopping"
	NodeStatusStopped      NodeStatus = "stopped"
	NodeStatusShuttingDown NodeStatus = "shutting_down"
	NodeStatusTerminated   NodeStatus = "terminated"
)

type NodeSpec struct {
	ProviderID ProviderID
	Extra      map[string]any
}

type Node struct {
	NodeID     NodeID
	ProviderID ProviderID
	Status     NodeStatus
	Addr       string
	Meta       map[string]any
	Cap        map[Cap]bool
}

func (n Node) ID() NodeID {
	return n.NodeID
}

func (n Node) HasCap(cap Cap) bool {
	return n.Cap[cap]
}

func (n *Node) SetCap(cap Cap, value bool) {
	n.Cap[cap] = value
}
