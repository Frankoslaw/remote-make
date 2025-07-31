package service

import (
	"fmt"
	"remoteMake/internal/model"
	"sync"
)

type NotificationManager struct {
	backends map[string]model.Notifier
	mu       sync.RWMutex
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		backends: make(map[string]model.Notifier),
	}
}

func (m *NotificationManager) RegisterNotifier(name string, n model.Notifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.backends[name] = n
}

func (m *NotificationManager) Notify(name string, subject, body string) error {
	m.mu.RLock()
	notifier, ok := m.backends[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("notifier %q not registered", name)
	}

	return notifier.Notify(subject, body)
}

func (m *NotificationManager) NotifyAll(subject, body string) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errors := make(map[string]error)
	for name, notifier := range m.backends {
		if err := notifier.Notify(subject, body); err != nil {
			errors[name] = err
		}
	}
	return errors
}
