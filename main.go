package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type TaskSpec struct {
	Name string
	// IsAtomic     bool
	// IsConcurrent bool
	InfraSpec InfraSpec
	StepsSpec []StepSpec
}

type InfraSpec struct {
	IsLocal bool
	// AwsSpec AwsSpec
}

type PosixSpec struct {
	Cmd   string
	Pwd   string
	Stdio string
}

type StepSpec struct {
	Task TaskSpec
	Proc PosixSpec
}

type Task struct {
	UUID        uuid.UUID
	Name        string
	State       string
	Infra       Infra
	StepResults []StepResult
}

type Infra struct {
	UUID  uuid.UUID
	State string
}

type StepResult struct {
	UUID       uuid.UUID
	State      string
	ProcResult ProcResult
}

type ProcResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type Envelope struct {
	TaskUUID       uuid.UUID       `json:"task_uuid"`
	MasterUUID     uuid.UUID       `json:"master_uuid"`
	WorkerUUID     uuid.UUID       `json:"worker_uuid"`
	ParentTaskUUID uuid.UUID       `json:"parent_task_uuid"`
	ParentStepUUID uuid.UUID       `json:"parent_step_uuid"`
	StepID         int             `json:"step_id"`
	Spec           json.RawMessage `json:"spec"`
}

var (
	tasks          = make(map[uuid.UUID]Task)
	nodeUUID       uuid.UUID
	ns             *server.Server
	nc             *nats.Conn
	self_terminate = true
)

func startNatsServer() *server.Server {
	opts := &server.Options{}
	srv, err := server.NewServer(opts)
	if err != nil {
		log.Fatalf("failed to create NATS server: %v", err)
	}
	go srv.Start()
	if !srv.ReadyForConnections(4 * time.Second) {
		log.Fatal("NATS server not ready for connections")
	}
	return srv
}

func connectToNats(srv *server.Server) *nats.Conn {
	conn, err := nats.Connect(srv.ClientURL())
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	return conn
}

func main() {
	nodeUUID = uuid.New()

	ns = startNatsServer()
	defer ns.Shutdown()

	nc = connectToNats(ns)
	defer nc.Close()

	// Master Events
	nc.Subscribe(fmt.Sprintf("node.%s.task.schedule", nodeUUID), func(task_spec_msg *nats.Msg) {
		log.Println(task_spec_msg.Subject)

		var task_spec TaskSpec
		_ = json.Unmarshal(task_spec_msg.Data, &task_spec)

		task_uuid := uuid.New()
		tasks[task_uuid] = Task{
			UUID:  task_uuid,
			Name:  task_spec.Name,
			State: "scheduled",
		}

		infra_spec := task_spec.InfraSpec
		var worker_uuid uuid.UUID
		if task_spec.InfraSpec.IsLocal {
			worker_uuid = nodeUUID
		} else {
			worker_uuid = uuid.New()
		}

		msgData, _ := json.Marshal(infra_spec)
		envelope := Envelope{
			TaskUUID:   task_uuid,
			MasterUUID: nodeUUID,
			WorkerUUID: worker_uuid,
			Spec:       msgData,
		}

		if infra_spec.IsLocal {
			envelope.WorkerUUID = nodeUUID
		} else {
			envelope.WorkerUUID = uuid.New()
		}

		msgData, _ = json.Marshal(envelope)

		sub, err := nc.SubscribeSync(fmt.Sprintf("node.%s.provisioned", worker_uuid))
		if err != nil {
			log.Fatal(err)
		}

		defer sub.Unsubscribe()
		nc.Publish(fmt.Sprintf("node.%s.provision", nodeUUID), msgData)

		infra_msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			log.Fatalf("provision timed out: %v", err)
		}
		log.Println(infra_msg.Subject)

		envelope.Spec = task_spec_msg.Data
		msgData, _ = json.Marshal(envelope)

		nc.Publish(fmt.Sprintf("node.%s.task.%s.start", worker_uuid, task_uuid), msgData)
	})

	nc.Subscribe(fmt.Sprintf("node.%s.provision", nodeUUID), func(infra_spec_msg *nats.Msg) {
		log.Println(infra_spec_msg.Subject)
		time.Sleep(time.Second)

		var envelope Envelope
		_ = json.Unmarshal(infra_spec_msg.Data, &envelope)

		var infra_spec InfraSpec
		_ = json.Unmarshal(envelope.Spec, &infra_spec)

		if infra_spec.IsLocal {
			infra := Infra{
				UUID:  envelope.WorkerUUID,
				State: "provisioned",
			}

			task := tasks[envelope.TaskUUID]
			task.Infra = infra
			tasks[envelope.TaskUUID] = task

			msgData, _ := json.Marshal(infra)
			nc.Publish(fmt.Sprintf("node.%s.provisioned", envelope.WorkerUUID), msgData)

			return
		}

		// TODO: Pulumi based shit + ssh worker init
	})

	nc.Subscribe("node.*.terminate", func(infra_msg *nats.Msg) {
		log.Println(infra_msg.Subject)

		var envelope Envelope
		json.Unmarshal(infra_msg.Data, &envelope)

		var infra Infra
		json.Unmarshal(envelope.Spec, &infra)

		if self_terminate && infra.UUID == nodeUUID {
			ns.Shutdown()
		}

		// TODO: Terminate remote instance
	})

	// Worker Events
	// Bind rest of worker events to this task
	nc.Subscribe(fmt.Sprintf("node.%s.task.*.start", nodeUUID), func(task_spec_msg *nats.Msg) {
		log.Println(task_spec_msg.Subject)

		var envelope Envelope
		json.Unmarshal(task_spec_msg.Data, &envelope)

		var task_spec TaskSpec
		json.Unmarshal(envelope.Spec, &task_spec)

		task := tasks[envelope.TaskUUID]

		// TODO: Support flags for IsAtomic and IsConcurent
		// TODO: Support AWS and cli interrupts
		for idx, step := range task_spec.StepsSpec {
			msgData, _ := json.Marshal(step)
			envelope.Spec = msgData
			envelope.StepID = idx
			msgData, _ = json.Marshal(envelope)

			sub, err := nc.SubscribeSync(fmt.Sprintf("node.%s.task.%s.step.%d.done", envelope.WorkerUUID, envelope.TaskUUID, idx))

			if err != nil {
				log.Fatal(err)
			}

			defer sub.Unsubscribe()
			nc.Publish(fmt.Sprintf("node.%s.task.%s.step.%d.start", envelope.WorkerUUID, envelope.TaskUUID, idx), msgData)

			step_result_msg, err := sub.NextMsg(5 * time.Second)
			if err != nil {
				log.Fatalf("step timed out: %v", err)
			}
			log.Println(step_result_msg.Subject)

			var step_result StepResult
			json.Unmarshal(step_result_msg.Data, &step_result)

			task.StepResults = append(task.StepResults, step_result)
		}

		tasks[envelope.TaskUUID] = task
		msgData, _ := json.Marshal(task)

		envelope.Spec = msgData

		msgData, _ = json.Marshal(envelope)

		nc.Publish(fmt.Sprintf("node.%s.task.%s.done", envelope.WorkerUUID, envelope.TaskUUID), msgData)
	})

	nc.Subscribe(fmt.Sprintf("node.%s.task.*.done", nodeUUID), func(task_result_msg *nats.Msg) {
		log.Println(task_result_msg.Subject)

		var envelope Envelope
		json.Unmarshal(task_result_msg.Data, &envelope)

		var task_result Task
		json.Unmarshal(envelope.Spec, &task_result)

		if envelope.ParentTaskUUID != uuid.Nil {
			// TODO: Propagate step result
		}

		msgData, _ := json.Marshal(&task_result.Infra)
		envelope.Spec = msgData
		msgData, _ = json.Marshal(envelope)

		nc.Publish(fmt.Sprintf("node.%s.terminate", envelope.WorkerUUID), msgData)
	})

	// TODO: Replace second wildcard with proper task id from the task start event
	// TODO: Use envelope to pass stdout, stderr and env if chainedEnv flag is enabled
	nc.Subscribe(fmt.Sprintf("node.%s.task.%s.step.*.start", nodeUUID, "*"), func(step_spec_msg *nats.Msg) {
		log.Println(step_spec_msg.Subject)

		var envelope Envelope
		json.Unmarshal(step_spec_msg.Data, &envelope)

		// TODO: Handle nested tasks
		// TODO: Handle notification and artifact managment steps
		// TODO: Do real work with posix processess]
		msgData, _ := json.Marshal(&StepResult{
			UUID:  uuid.New(),
			State: "done",
		})

		nc.Publish(fmt.Sprintf("node.%s.task.%s.step.%d.done", envelope.WorkerUUID, envelope.TaskUUID, envelope.StepID), msgData)
	})

	msgData, _ := json.Marshal(TaskSpec{
		Name: "Test task",
		InfraSpec: InfraSpec{
			IsLocal: true,
		},
		StepsSpec: []StepSpec{
			{
				Proc: PosixSpec{
					Cmd: "echo 'Hello world'",
				},
			},
			{
				Proc: PosixSpec{
					Cmd: "echo 'Hello world 2'",
				},
			},
		},
	})

	nc.Publish(fmt.Sprintf("node.%s.task.schedule", nodeUUID), msgData)
	ns.WaitForShutdown()
}
