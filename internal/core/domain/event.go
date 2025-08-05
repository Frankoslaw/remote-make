package domain

const (
	// Commands (Requests)
	EventTaskStart     = "node.%s.task.start"
	EventStepStart     = "node.%s.step.start"
	EventNodeProvision = "node.%s.node.provision"
	EventNodeTerminate = "node.%s.node.terminate"

	// Status Broadcasts
	EventNodeReady       = "node.%s.ready"
	EventNodeInterrupted = "node.%s.interrupted"
)
