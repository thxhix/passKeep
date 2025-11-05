package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/thxhix/passKeeper/internal/config"
)

// PostgresStorage wraps a *sql.DB driver used by repository implementations.
type PostgresStorage struct {
	Driver *sql.DB
}

// NewPostgres opens a connection to Postgres using config.PostgresQL and verifies
// the connection by calling PingContext with the provided ctx.
//
// The returned *PostgresStorage should be closed by calling Driver.Close()
// (for example via a close function returned by the storage wiring).
func NewPostgres(ctx context.Context, config *config.Config) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", config.PostgresQL)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PostgresStorage{
		Driver: db,
	}, nil
}
