package posix

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"remoteMake/internal/model"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type posixProcess struct {
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout, stderr io.ReadCloser
	done           chan model.State
	once           sync.Once
	mu             sync.Mutex
	exitCallbacks  []func(model.State)
}

func (p *posixProcess) ID() string            { return strconv.Itoa(p.cmd.Process.Pid) }
func (p *posixProcess) Stdin() io.WriteCloser { return p.stdin }
func (p *posixProcess) Stdout() io.Reader     { return p.stdout }
func (p *posixProcess) Stderr() io.Reader     { return p.stderr }

func (p *posixProcess) Start(ctx context.Context) (<-chan model.State, error) {
	if err := p.cmd.Start(); err != nil {
		p.once.Do(func() {
			state := model.State{ExitCode: -1, Err: err}
			p.invokeCallbacks(state)
			p.done <- state
			close(p.done)
		})
		return p.done, err
	}

	go func() {
		defer p.once.Do(func() { close(p.done) })

		waitCh := make(chan model.State, 1)
		go func() {
			st, err := p.cmd.Process.Wait()
			waitCh <- model.State{ExitCode: st.ExitCode(), Err: err}
		}()

		select {
		case <-ctx.Done():
			_ = p.Stop()
			state := <-waitCh
			p.invokeCallbacks(state)
			p.done <- state
		case state := <-waitCh:
			p.invokeCallbacks(state)
			p.done <- state
		}
	}()

	return p.done, nil
}

func (p *posixProcess) Wait() (model.State, error) {
	state, ok := <-p.done
	if !ok {
		return model.State{}, errors.New("process channel closed before state received")
	}

	for range p.done {
	}

	return state, state.Err
}

func (p *posixProcess) Stop() error {
	if p.cmd.Process == nil {
		return errors.New("not started")
	}
	_ = p.cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-time.After(time.Second):
		return p.cmd.Process.Kill()
	case <-p.done:
		return nil
	}
}

func (p *posixProcess) OnExit(cb func(model.State)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.exitCallbacks = append(p.exitCallbacks, cb)
}

func (p *posixProcess) invokeCallbacks(state model.State) {
	p.mu.Lock()
	cbs := make([]func(model.State), len(p.exitCallbacks))
	copy(cbs, p.exitCallbacks)
	p.mu.Unlock()
	for _, cb := range cbs {
		cb(state)
	}
}
