package storage

import (
	"context"
	"github.com/thxhix/passKeeper/internal/adapters/storage/postgres"
	"github.com/thxhix/passKeeper/internal/config"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"github.com/thxhix/passKeeper/internal/domain/user"
	"go.uber.org/zap"
	"time"
)

// Storage groups repository instances used by the application.
type Storage struct {
	User     user.UserRepository
	Token    token.TokenRepository
	Keychain keychain.KeychainRepository
}

// NewStorage creates repository instances, runs migrations and returns a cleanup
// function. It expects cfg.PostgresQL to be set to a valid DSN. If cfg.PostgresQL
// is empty ErrNoPostgresConnection is returned.
//
// The returned close function should be called to close DB connections when the
// application shuts down.
//
// Example:
//
//	storage, closeFn, err := storage.NewStorage(ctx, cfg, logger)
//	if err != nil { return err }
//	defer closeFn()
func NewStorage(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Storage, func(), error) {
	if cfg.PostgresQL == "" {
		return nil, nil, ErrNoPostgresConnection
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.DatabaseInitTimeoutSeconds)*time.Second)
	defer cancel()

	logger.Info("Trying to connect to postgresql", zap.String("DSN", cfg.PostgresQL))
	db, err := postgres.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	logger.Info("Trying to migrate", zap.String("migrations_path", cfg.MigrationsPath))
	migrator := postgres.NewMigrator(db.Driver, cfg.MigrationsPath)
	if err := migrator.Up(); err != nil {
		return nil, nil, err
	}

	closeFn := func() { _ = db.Driver.Close() }

	userRepository := postgres.NewUsersRepository(db.Driver)
	tokenRepository := postgres.NewTokensRepository(db.Driver)
	keychainRepository := postgres.NewKeychainRepository(db.Driver)

	return &Storage{
		User:     userRepository,
		Token:    tokenRepository,
		Keychain: keychainRepository,
	}, closeFn, nil
}
