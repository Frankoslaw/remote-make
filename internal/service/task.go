package service

import (
	"fmt"
	"remoteMake/internal/model"
	"sync"
)

type TaskUID = int

type Task struct {
	UID  TaskUID
	Proc model.Process
}

type TaskManager struct {
	uidGen   *UIDService
	backends map[string]model.Runner
	tasks    map[TaskUID]Task
	mu       sync.RWMutex
}

func NewTaskManager(uidGen *UIDService) *TaskManager {
	return &TaskManager{
		uidGen:   uidGen,
		backends: make(map[string]model.Runner),
		tasks:    make(map[TaskUID]Task),
	}
}

func (m *TaskManager) RegisterRunner(name string, r model.Runner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.backends[name] = r
}

func (m *TaskManager) CreateTask(runnerName string, spec model.Spec) (Task, error) {
	m.mu.RLock()
	r, ok := m.backends[runnerName]
	m.mu.RUnlock()
	if !ok {
		return Task{}, fmt.Errorf("runner %q not found", runnerName)
	}

	proc, err := r.Create(spec)
	if err != nil {
		return Task{}, err
	}

	uid := m.uidGen.Generate()
	task := Task{
		UID:  uid,
		Proc: proc,
	}

	m.mu.Lock()
	m.tasks[uid] = task
	m.mu.Unlock()

	return task, nil
}

func (m *TaskManager) GetTask(id TaskUID) (Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[id]
	return task, ok
}
