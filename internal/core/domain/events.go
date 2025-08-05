package domain

const (
	EventNodeReady       = "node.%s.ready"
	EventNodeInterrupted = "node.%s.interrupted"
	EventStepStart       = "node.%s.step.%s.start"
	EventStepDone        = "step.%s.done"
	EventStepError       = "step.%s.error"
)
