package domain

type WorkerState int
type TaskState int
type StepState int
type ProcessState int

const (
	WorkerError     WorkerState = 0
	WorkerScheduled             = iota + 1
	WorkerProvisioning
	WorkerProvisioned
	WorkerTerminating
	WorkerTerminated
)

const (
	TaskError     TaskState = 0
	TaskScheduled           = iota + 10
	TaskRunning
	TaskDone
)

const (
	StepError     StepState = 0
	StepScheduled           = iota + 20
	StepRunning
	StepDone
)

const (
	ProcessError     ProcessState = 0
	ProcessScheduled              = iota + 30
	ProcessRunning
	ProcessDone
)
