package http_server

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/thxhix/passKeeper/internal/config"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server defines an interface for starting the HTTP server.
type Server interface {
	Start() error
}

// server implements the Server interface.
type server struct {
	http   *http.Server
	router *chi.Mux
	cfg    *config.Config
	logger *zap.Logger
}

// NewServer creates a new Server instance with the provided router, config, and logger.
// The returned Server is ready to start with the Start() method.
func NewServer(router *chi.Mux, cfg *config.Config, logger *zap.Logger) Server {
	srv := &http.Server{
		Addr:    cfg.RESTAddress,
		Handler: router,
	}

	return &server{
		http:   srv,
		router: router,
		cfg:    cfg,
		logger: logger,
	}
}

// Start runs the HTTP server in a separate goroutine and waits for either:
// 1. A server error from ListenAndServe (returns that error), or
// 2. A termination signal (SIGINT, SIGTERM, SIGQUIT) to gracefully shutdown the server.
// It returns an error if the shutdown fails.
func (s *server) Start() error {
	s.logger.Info("HTTP server startup", zap.String("address", s.cfg.RESTAddress))

	errCh := make(chan error, 1)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		return err

	case sig := <-sigCh:
		s.logger.Info("HTTP server graceful shutdown by signal", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.http.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server error", zap.Error(err))
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	}
}
