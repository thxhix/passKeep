package storage

import (
	"context"
	"github.com/thxhix/passKeeper/internal/config"
	"go.uber.org/zap"
	"testing"
)

func TestNewStorage_NoDSN_ReturnsError(t *testing.T) {
	ctx := context.Background()

	// minimal config with empty PostgresQL
	cfg := &config.Config{
		PostgresQL:                 "",
		DatabaseInitTimeoutSeconds: 1,
		MigrationsPath:             "",
	}

	logger := zap.NewNop()

	_, closeFn, err := NewStorage(ctx, cfg, logger)
	if closeFn != nil {
		// should not return a close function on error, but if it does, call it
		defer closeFn()
	}
	if err == nil {
		t.Fatalf("expected error when PostgresDSN is empty, got nil")
	}

	// If you exported ErrNoPostgresConnection, assert equality:
	if err != ErrNoPostgresConnection {
		t.Fatalf("expected ErrNoPostgresConnection, got: %v", err)
	}
}
