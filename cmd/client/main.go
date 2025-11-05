package main

import (
	"github.com/thxhix/passKeeper/internal/app"
	"github.com/thxhix/passKeeper/internal/config"
	customLogger "github.com/thxhix/passKeeper/internal/logger"
	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v1"
	"log"
)

func main() {
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := customLogger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	cliApp := cli.NewApp()
	cliApp.Name = "passKeeper"
	cliApp.Version = "dev"

	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "base-url, b",
			Usage:  "API base URL",
			Value:  cfg.ServerAddress,
			EnvVar: "GK_BASE_URL",
		},
	}

	err = app.RunClient(cfg, logger, cliApp)
	if err != nil {
		logger.Fatal("Client startup critical error", zap.Error(err))
	}

	return
}
