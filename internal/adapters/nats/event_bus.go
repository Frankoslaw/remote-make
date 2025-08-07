package nats

import (
	"context"
	"encoding/json"
	"remote-make/internal/adapters/config"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type EmbeddedNats struct {
	server *server.Server
}

func NewEmbeddedNats() (*EmbeddedNats, error) {
	opts := &server.Options{}
	ns, err := server.NewServer(opts)

	if err != nil {
		return nil, err
	}

	go ns.Start()

	if !ns.ReadyForConnections(4 * time.Second) {
		return nil, err
	}

	return &EmbeddedNats{server: ns}, nil
}

func (n *EmbeddedNats) ClientURL() string {
	return n.server.ClientURL()
}

func (e *EmbeddedNats) Shutdown() {
	if e.server != nil {
		e.server.Shutdown()
	}
}

type NatsEventBus struct {
	conn *nats.Conn
}

func NewNatsEventBus(cfg *config.Config) (*NatsEventBus, error) {
	nc, err := nats.Connect(cfg.NATSURL)

	if err != nil {
		return nil, err
	}

	return &NatsEventBus{
		conn: nc,
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
}

func MustMarshal[T any](step T) []byte {
	data, err := json.Marshal(step)
	if err != nil {
		panic(err)
	}
	return data
}

func MustUnmarshal[T any](data []byte) T {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		panic(err)
	}
	return result
}
