package app

import (
	"context"
	reposStorage "github.com/thxhix/passKeeper/internal/adapters/storage"
	"github.com/thxhix/passKeeper/internal/client/api"
	"github.com/thxhix/passKeeper/internal/client/cli/commands"
	"github.com/thxhix/passKeeper/internal/client/client_services"
	"github.com/thxhix/passKeeper/internal/config"
	"github.com/thxhix/passKeeper/internal/security"
	"github.com/thxhix/passKeeper/internal/server/http_server"
	"github.com/thxhix/passKeeper/internal/services"
	"github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http"
	"github.com/thxhix/passKeeper/internal/transport/http/handlers"
	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func RunServer(cfg *config.Config, logger *zap.Logger) error {
	var err error

	ctx := context.Background()

	storage, closeFn, err := reposStorage.NewStorage(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to create repo storage", zap.Error(err))
		return err
	}

	defer closeFn()

	hasher := security.NewHasher()
	jwtManager := security.NewJWTManager(cfg)
	authService := services.NewAuthService(storage.User, storage.Token, &hasher, &jwtManager)

	aead, err := security.NewAEAD(logger, cfg)
	keychainService := services.NewKeychainService(storage.Keychain, aead)

	h := handlers.NewHandlers(logger, &authService, &keychainService)
	r := http.NewRouter(h, &jwtManager)
	s := http_server.NewServer(r, cfg, logger)

	err = s.Start()
	if err != nil {
		logger.Error("Failed to start http server", zap.Error(err))
		return err
	}

	return err
}

func RunClient(cfg *config.ClientConfig, logger *zap.Logger, cliApp *cli.App) error {
	var err error

	httpClient, err := client_http.NewHttpClient(cfg.ServerAddress, logger)
	if err != nil {
		logger.Error("Failed to create http client", zap.Error(err))
		return err
	}

	authAPI := api.NewAuthAPI(httpClient)
	authService := client_services.NewAuthClientService(authAPI, httpClient)
	authCmd := commands.NewAuthCLICommands(authService)

	keychainAPI := api.NewKeychainAPI(httpClient)
	keychainService := client_services.NewKeychainClientService(keychainAPI, httpClient)
	keychainCmd := commands.NewKeychainCLICommands(keychainService)

	cliApp.Commands = []cli.Command{
		authCmd.RegisterCmd(),
		authCmd.LoginCmd(),
		authCmd.RefreshTokenCmd(),

		keychainCmd.Add(),
		keychainCmd.List(),
		keychainCmd.Get(),
		keychainCmd.Delete(),
	}

	return cliApp.Run(os.Args)
}
