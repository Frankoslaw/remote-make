package model

type Notifier interface {
	Notify(subject string, body string) error
}
