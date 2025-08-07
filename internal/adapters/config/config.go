package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Config struct {
	NodeUUID uuid.UUID

	TUIEnabled bool
	APIEnabled bool

	NodeManagerEnabled bool
	TaskRunnerEnabled  bool
	StepRunnerEnabled  bool

	NodeManagerTimeout int
	TaskRunnerTimeout  int
	StepRunnerTimeout  int

	EmbeddedNATSEnabled bool
	NATSURL             string

	EmbeddedSQLiteEnabled bool
	SQLiteURL             string
}

func Load() *Config {
	godotenv.Load()

	var err error
	cfg := &Config{}

	// Required
	nodeUUIDstr := os.Getenv("NODE_UUID")
	if nodeUUIDstr == "" {
		cfg.NodeUUID = uuid.New()
		slog.Warn("NODE_UUID is not set, generating a new UUID")
	} else {
		cfg.NodeUUID, err = uuid.Parse(nodeUUIDstr)
		if err != nil {
			panic("Invalid NODE_UUID format: " + err.Error())
		}
	}

	cfg.TUIEnabled, _ = strconv.ParseBool(os.Getenv("TUI_ENABLED"))
	cfg.APIEnabled, _ = strconv.ParseBool(os.Getenv("REST_API_ENABLED"))
	cfg.NodeManagerEnabled, _ = strconv.ParseBool(os.Getenv("NODE_MANAGER_ENABLED"))
	cfg.TaskRunnerEnabled, _ = strconv.ParseBool(os.Getenv("TASK_RUNNER_ENABLED"))
	cfg.StepRunnerEnabled, _ = strconv.ParseBool(os.Getenv("STEP_RUNNER_ENABLED"))

	// Timeouts (default -1)
	cfg.NodeManagerTimeout = parseIntOrDefault("NODE_MANAGER_TIMEOUT", -1)
	cfg.TaskRunnerTimeout = parseIntOrDefault("TASK_RUNNER_TIMEOUT", -1)
	cfg.StepRunnerTimeout = parseIntOrDefault("STEP_RUNNER_TIMEOUT", -1)

	// NATS
	natsEmbeddedEnv := os.Getenv("EMBEDDED_NATS_ENABLED")
	if natsEmbeddedEnv == "" {
		cfg.EmbeddedNATSEnabled = true
		slog.Warn("EMBEDDED_NATS_ENABLED is not set, defaulting to true")
	} else {
		cfg.EmbeddedNATSEnabled, _ = strconv.ParseBool(natsEmbeddedEnv)
	}
	cfg.NATSURL = os.Getenv("NATS_URL")
	if cfg.NATSURL == "" {
		cfg.NATSURL = "nats://localhost:4222"
	}

	// SQLite
	sqliteEmbeddedEnv := os.Getenv("EMBEDDED_SQLITE_ENABLED")
	if sqliteEmbeddedEnv == "" {
		cfg.EmbeddedSQLiteEnabled = true
		slog.Warn("EMBEDDED_SQLITE_ENABLED is not set, defaulting to true")
	} else {
		cfg.EmbeddedSQLiteEnabled, _ = strconv.ParseBool(sqliteEmbeddedEnv)
	}
	cfg.SQLiteURL = os.Getenv("SQLITE_URL")
	if cfg.SQLiteURL == "" {
		cfg.SQLiteURL = "file:./remote-make.db"
	}

	return cfg
}

func parseIntOrDefault(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return i
}
