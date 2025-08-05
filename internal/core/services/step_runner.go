package services

import (
	"encoding/json"
	"fmt"
	"remote-make/internal/core/domain"
	"remote-make/internal/core/ports"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type StepRunner struct {
	nodeIDRepo ports.NodeIdentityRepo
	eventBus   ports.EventBus
	procRunner ports.ProcessRunner
}

func NewStepRunner(ni ports.NodeIdentityRepo, ev ports.EventBus, pr ports.ProcessRunner) *StepRunner {
	nodeID := ni.NodeUUID()
	r := &StepRunner{nodeIDRepo: ni, eventBus: ev, procRunner: pr}

	ev.Subscribe(fmt.Sprintf(domain.EventStepStart, nodeID, "*"), func(msg *nats.Msg) {
		stepID, _ := uuid.Parse(strings.Split(msg.Subject, ".")[3])

		var st domain.StepTemplate
		json.Unmarshal(msg.Data, &st)

		step, err := r.Start(st)
		msgData, _ := json.Marshal(&step)

		if err != nil {
			ev.Publish(fmt.Sprintf(domain.EventStepError, stepID), msgData)
			return
		}

		ev.Publish(fmt.Sprintf(domain.EventStepDone, stepID), msgData)
	})

	return r
}

func (r *StepRunner) Start(st domain.StepTemplate) (domain.Step, error) {
	res, err := r.procRunner.Start(st.ProcessTemplate)
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
