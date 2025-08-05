package domain

type WorkerState int
type TaskState int
type StepState int

const (
	WorkerScheduled WorkerState = iota + 100
	WorkerProvisioning
	WorkerProvisioned
	WorkerTerminating
	WorkerTerminated
	WorkerError
)

const (
	TaskScheduled TaskState = iota + 200
	TaskRunning
	TaskDone
	TaskError
)

const (
	StepScheduled StepState = iota + 300
	StepRunning
	StepDone
	StepError
)
