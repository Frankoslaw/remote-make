package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type NatsEventBus struct {
	conn   *nats.Conn
	server *server.Server
}

func NewEmbeddedNatsEventBus() (*NatsEventBus, error) {
	opts := &server.Options{}
	ns, err := server.NewServer(opts)

	if err != nil {
		return nil, err
	}

	go ns.Start()

	if !ns.ReadyForConnections(4 * time.Second) {
		return nil, err
	}

	nc, err := nats.Connect(ns.ClientURL())

	if err != nil {
		ns.Shutdown()
		return nil, err
	}

	return &NatsEventBus{
		conn:   nc,
		server: ns,
	}, nil
}

func (n *NatsEventBus) Publish(subject string, data []byte) error {
	return n.conn.Publish(subject, data)
}

func (n *NatsEventBus) Subscribe(subject string, handler func(msg *nats.Msg)) error {
	_, err := n.conn.Subscribe(subject, handler)
	return err
}

func (n *NatsEventBus) Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error) {
	return n.conn.RequestWithContext(ctx, subject, data)
}

func (n *NatsEventBus) Shutdown() {
	if n.conn != nil {
		n.conn.Close()
	}
	if n.server != nil {
		n.server.Shutdown()
	}
}

func marshal[T any](step T) []byte {
	data, _ := json.Marshal(step)
	return data
}

type NodeManagerHandler struct {
	nodeIDRepo  ports.NodeIdentityRepo
	nodeManager ports.NodeManager
}

func NewNodeManagerHandler(ni ports.NodeIdentityRepo, nm ports.NodeManager) *NodeManagerHandler {
	return &NodeManagerHandler{nodeIDRepo: ni, nodeManager: nm}
}

func (h *NodeManagerHandler) RegisterSubs(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventNodeProvision, h.nodeIDRepo.NodeUUID()), h.NodeProvision)
	ev.Subscribe(fmt.Sprintf(domain.EventNodeTerminate, h.nodeIDRepo.NodeUUID()), h.NodeTerminate)
}

func (h *NodeManagerHandler) NodeProvision(msg *nats.Msg) {
	var tmpl domain.WorkerTemplate
	err := json.Unmarshal(msg.Data, &tmpl)
	if err != nil {
		errorWorker := domain.Worker{ID: tmpl.ID, State: domain.WorkerError}
		_ = msg.Respond(marshal(errorWorker))
		return
	}

	// TODO: Hardcoded time
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	worker, _ := h.nodeManager.Provision(ctx, tmpl)

	_ = msg.Respond(marshal(worker))
}

func (h *NodeManagerHandler) NodeTerminate(msg *nats.Msg) {
	var tmpl domain.Worker
	err := json.Unmarshal(msg.Data, &tmpl)
	if err != nil {
		errorWorker := domain.Worker{ID: tmpl.ID, State: domain.WorkerError}
		_ = msg.Respond(marshal(errorWorker))
		return
	}

	// TODO: Hardcoded time
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	worker, _ := h.nodeManager.Terminate(ctx, tmpl)

	_ = msg.Respond(marshal(worker))
}

type TaskRunnerHandler struct {
	nodeIDRepo ports.NodeIdentityRepo
	taskRunner ports.TaskRunner
}

func NewTaskRunnerHandler(ni ports.NodeIdentityRepo, tr ports.TaskRunner) *TaskRunnerHandler {
	return &TaskRunnerHandler{nodeIDRepo: ni, taskRunner: tr}
}

func (h *TaskRunnerHandler) RegisterSubs(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventTaskStart, h.nodeIDRepo.NodeUUID()), h.TaskStart)
}

func (h *TaskRunnerHandler) TaskStart(msg *nats.Msg) {
	var tmpl domain.TaskTemplate
	err := json.Unmarshal(msg.Data, &tmpl)
	if err != nil {
		errorTask := domain.Task{ID: tmpl.ID, State: domain.TaskError}
		_ = msg.Respond(marshal(errorTask))
		return
	}

	// TODO: Hardcoded time
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	task, _ := h.taskRunner.Start(ctx, tmpl)

	_ = msg.Respond(marshal(task))
}

type StepRunnerHandler struct {
	nodeIDRepo ports.NodeIdentityRepo
	stepRunner ports.StepRunner
}

func NewStepRunnerHandler(ni ports.NodeIdentityRepo, sr ports.StepRunner) *StepRunnerHandler {
	return &StepRunnerHandler{nodeIDRepo: ni, stepRunner: sr}
}

func (h *StepRunnerHandler) RegisterSubs(ev ports.EventBus) {
	ev.Subscribe(fmt.Sprintf(domain.EventStepStart, h.nodeIDRepo.NodeUUID()), h.StepStart)
}

func (h *StepRunnerHandler) StepStart(msg *nats.Msg) {
	var tmpl domain.StepTemplate
	err := json.Unmarshal(msg.Data, &tmpl)
	if err != nil {
		errorStep := domain.Step{ID: tmpl.ID, State: domain.StepError}
		_ = msg.Respond(marshal(errorStep))
		return
	}

	// TODO: Hardcoded time
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	step, _ := h.stepRunner.Start(ctx, tmpl)

	_ = msg.Respond(marshal(step))
}
