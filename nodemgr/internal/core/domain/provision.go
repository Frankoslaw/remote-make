package domain

type NodeID string
type ProviderID string
type Cap string

type NodeSpec struct {
	ProviderID ProviderID
	Extra      map[string]any
}

type Node struct {
	NodeID     NodeID
	ProviderID ProviderID

	State NodeState
	Meta  map[string]any
	Cap   map[Cap]bool
}

func (n Node) ID() NodeID {
	return n.NodeID
}

func (n *Node) HasCap(cap Cap) bool {
	return n.Cap[cap]
}

func (n *Node) SetCap(cap Cap, value bool) {
	n.Cap[cap] = value
}
