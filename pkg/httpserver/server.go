package httpserver

import (
	"context"

	"net/http"
	"time"

	"github.com/menyasosali/mts/pkg/logger"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":8085"
	_defaultShutdownTimeout = 3 * time.Second
)

type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewServer(ctx context.Context, logger logger.Interface, handler http.Handler, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.start(ctx, logger)

	return s
}

func (s *Server) start(ctx context.Context, logger logger.Interface) {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", err)
			s.notify <- err
			return
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down HTTP server...")
			ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
			defer cancel()

			err := s.server.Shutdown(ctx)
			if err != nil {
				logger.Error("HTTP server shutdown error", err)
				s.notify <- err
			} else {
				logger.Info("HTTP server stopped")
				s.notify <- nil
			}
		}
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
