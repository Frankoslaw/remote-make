package infra

import (
	"context"

	"github.com/containers/podman/v5/pkg/bindings"
)

func ConnectPodman(url string) (context.Context, error) {
	return bindings.NewConnection(context.Background(), url)
}
