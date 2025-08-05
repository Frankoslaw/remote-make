package adapters

import (
	"context"
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

// TODO: nats.Msg should be abstracted away
func (n *NatsEventBus) Subscribe(subject string, handler func(m *nats.Msg)) error {
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
