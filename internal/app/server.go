package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/karzhen/restaurant-lk/internal/config"
)

type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewHTTPServer(cfg config.Config, handler http.Handler, logger *slog.Logger) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:         cfg.HTTPAddr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout(),
			WriteTimeout: cfg.WriteTimeout(),
			IdleTimeout:  cfg.IdleTimeout(),
		},
		logger: logger,
	}
}

func (s *HTTPServer) Start() error {
	s.logger.Info("starting http server", "addr", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}
