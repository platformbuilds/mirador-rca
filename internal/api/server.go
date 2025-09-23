package api

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/miradorstack/mirador-rca/internal/config"
	rcav1 "github.com/miradorstack/mirador-rca/internal/grpc/generated"
)

// Server wraps the gRPC server implementation and lifecycle helpers.
type Server struct {
	cfg        config.ServerConfig
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer constructs a gRPC server bound to the configured address.
func NewServer(cfg config.ServerConfig, service rcav1.RCAEngineServer, opts ...grpc.ServerOption) (*Server, error) {
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", cfg.Address, err)
	}

	grpcServer := grpc.NewServer(opts...)

	rcav1.RegisterRCAEngineServer(grpcServer, service)

	// Register health service so probes can hit /health via gRPC reflection tools.
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	// Enable server reflection in development environments.
	reflection.Register(grpcServer)

	return &Server{
		cfg:        cfg,
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

// Start serves incoming gRPC requests until Stop/Shutdown is invoked.
func (s *Server) Start() error {
	if s.grpcServer == nil || s.listener == nil {
		return fmt.Errorf("server not initialised")
	}
	return s.grpcServer.Serve(s.listener)
}

// Shutdown attempts a graceful shutdown, falling back to Stop after timeout.
func (s *Server) Shutdown(ctx context.Context) {
	if s.grpcServer == nil {
		return
	}

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
	case <-stopped:
	}
}

// Address exposes the bound listener address (useful for tests).
func (s *Server) Address() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// GracefulTimeout returns the configured graceful timeout duration.
func (s *Server) GracefulTimeout() time.Duration {
	return s.cfg.GracefulTimeout
}
