package main

import (
	"github.com/thxhix/passKeeper/internal/app"
	"github.com/thxhix/passKeeper/internal/config"
	customLogger "github.com/thxhix/passKeeper/internal/logger"
	"go.uber.org/zap"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := customLogger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	err = app.RunServer(cfg, logger)
	if err != nil {
		logger.Fatal("App startup critical error", zap.Error(err))
	}
}
