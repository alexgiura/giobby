package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(port string, handler http.Handler) (*Server, error) {
	if port == "" {
		return nil, fmt.Errorf("server port is empty")
	}
	if handler == nil {
		return nil, fmt.Errorf("http handler is nil")
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%s", port),
			Handler:           handler,
			ReadTimeout:       10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20, // 1 MiB
		},
	}, nil
}

func (s *Server) Start() error {
	if s == nil || s.httpServer == nil {
		return fmt.Errorf("http server is nil")
	}

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.httpServer == nil {
		return fmt.Errorf("http server is nil")
	}
	if ctx == nil {
		return fmt.Errorf("shutdown context is nil")
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	return nil
}
