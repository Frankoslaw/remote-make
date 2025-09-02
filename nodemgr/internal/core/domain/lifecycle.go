package domain

type NodeState string

const (
	NodeStatePending      NodeState = "pending"
	NodeStateRunning      NodeState = "running"
	NodeStateStopping     NodeState = "stopping"
	NodeStateStopped      NodeState = "stopped"
	NodeStateShuttingDown NodeState = "shutting_down"
	NodeStateTerminated   NodeState = "terminated"
)
