package domain

type NodeStatus string

const (
	NodeStatusPending    NodeStatus = "pending"
	NodeStatusRunning    NodeStatus = "running"
	NodeStatusStopped    NodeStatus = "stopped"
	NodeStatusFailed     NodeStatus = "failed"
	NodeStatusDestroying NodeStatus = "destroying"
)

type Node struct {
	NodeID   string     `json:"id,omitempty"`
	Status   NodeStatus `json:"status"`
	Provider string     `json:"provider"`

	Conn NodeConn          `json:"conn,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
	Caps map[string]bool   `json:"caps,omitempty"`
}

func (n *Node) ID() string {
	return n.NodeID
}

func (n *Node) Node() *Node {
	return n
}

func (n *Node) Tag(key string) (string, bool) {
	if n == nil || n.Tags == nil {
		return "", false
	}
	v, ok := n.Tags[key]
	return v, ok
}

func (n *Node) SetTag(key, value string) {
	if n == nil {
		return
	}
	if n.Tags == nil {
		n.Tags = make(map[string]string, 4)
	}
	n.Tags[key] = value
}

func (n *Node) HasCap(key string) bool {
	if n == nil || n.Caps == nil {
		return false
	}
	return n.Caps[key]
}

func (n *Node) SetCap(key string, value bool) {
	if n == nil {
		return
	}
	if n.Caps == nil {
		n.Caps = make(map[string]bool, 4)
	}
	n.Caps[key] = value
}

type NodeConn struct {
	Hostname  string `json:"hostname,omitempty"`
	PublicIP  string `json:"public_ip,omitempty"`
	PrivateIP string `json:"private_ip,omitempty"`

	SSHUser string `json:"ssh_user,omitempty"`
	SSHPass string `json:"ssh_pass,omitempty"`
	SSHPort int    `json:"ssh_port,omitempty"`
	SSHKey  string `json:"ssh_key,omitempty"`
}
