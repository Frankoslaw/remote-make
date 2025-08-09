//go:build worker
// +build worker

package main

import (
	"log/slog"
	"os"
	"remote-make/internal/adapters"
	"remote-make/internal/adapters/config"
	"remote-make/internal/adapters/nats"
	"remote-make/internal/core/services"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	// Generic services
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	cfg := config.Load()
	slog.Info("Configuration loaded", "config", cfg)

	eventBus, err := nats.NewNatsEventBus(cfg)
	if err != nil {
		slog.Error("Failed to connect to NATS server", "error", err)
		return
	}
	defer eventBus.Shutdown()
	slog.Info("Connected to NATS server", "url", cfg.NATSURL)

	// Worker services
	healthSubscriber := nats.NewHealthSubscriber(cfg.NodeID)
	healthSubscriber.RegisterSubscribers(eventBus)

	procRunner := adapters.NewLocalProcessRunner()
	stepRunner := services.NewStepRunner(cfg.NodeID, eventBus, procRunner)
	stepRunnerSubscriber := nats.NewStepRunnerSubscriber(cfg.NodeID, stepRunner)
	stepRunnerSubscriber.RegisterSubscribers(eventBus)

	// keep alive server
	select {}
}
