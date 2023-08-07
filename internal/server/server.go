package server

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/menyasosali/mts/internal/server/gateway"
	pb "github.com/menyasosali/mts/pkg/gen"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strings"
	"time"

	"github.com/menyasosali/mts/pkg/logger"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":8080"
	_defaultShutdownTimeout = 3 * time.Second
)

type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewServer(ctx context.Context, log logger.Interface, service *gateway.Service, opts ...Option) *Server {
	grpcServer := grpc.NewServer()
	grpcOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	gwmux := runtime.NewServeMux()
	pb.RegisterGatewayServer(grpcServer, service)
	gwmux.HandlePath("POST", "/images/upload", service.UploadImageHandler)

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	httpServer := &http.Server{
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
		Handler:      grpcHandlerFunc(grpcServer, mux),
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	pb.RegisterGatewayHandlerFromEndpoint(ctx, gwmux, httpServer.Addr, grpcOpts)

	s.start(ctx, log)

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

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
