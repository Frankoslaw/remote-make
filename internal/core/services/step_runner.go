package services

import (
	"context"
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type StepRunner struct {
	nodeIDRepo  ports.NodeIdentityRepo
	eventBus    ports.EventBus
	procRunner  ports.ProcessRunner
	stepTimeout time.Duration
}

func NewStepRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, pr ports.ProcessRunner) *StepRunner {
	nodeID := ni.NodeUUID()

	r := &StepRunner{nodeIDRepo: ni, eventBus: ev, procRunner: pr, stepTimeout: 5 * time.Second}
	ev.Subscribe(fmt.Sprintf(domain.EventStepStart, nodeID, "*"), r.handleStart)

	return r
}

func (r *StepRunner) handleStart(msg *nats.Msg) {
	var tmpl domain.StepTemplate
	err := json.Unmarshal(msg.Data, &tmpl)
	if err != nil {
		errorStep := domain.Step{ID: tmpl.ID, State: domain.StepError}
		_ = msg.Respond(marshal(errorStep))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.stepTimeout)
	defer cancel()
	step, _ := r.Start(ctx, tmpl)

	_ = msg.Respond(marshal(step))
}

func (r *StepRunner) Start(ctx context.Context, st domain.StepTemplate) (domain.Step, error) {
	res, err := r.procRunner.Start(ctx, st.ProcessTemplate)
	step := domain.Step{
		ID:            uuid.New(),
		ProcessResult: res,
	}

	if err != nil {
		step.State = domain.StepError
	} else {
		step.State = domain.StepDone
	}

	return step, err
}

func marshal(step domain.Step) []byte {
	data, _ := json.Marshal(step)
	return data
}
