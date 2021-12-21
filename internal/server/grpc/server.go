package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/sashayakovtseva/hello-web/internal/config"
	api "github.com/sashayakovtseva/hello-web/pkg/grpc"
)

// Server knows how to serve gRPC requests.
type Server struct {
	api.UnimplementedHelloServiceServer

	logger *zap.Logger
	config *config.Server
	health *health.Server
}

// NewServer returns new Server object that will use passed config
// during setup. To start serving requests call Server.Serve.
func NewServer(
	logger *zap.Logger,
	config *config.Server,
) *Server {
	grpcZap.ReplaceGrpcLoggerV2(logger)

	return &Server{
		logger: logger,
		config: config,
		health: health.NewServer(),
	}
}

// Serve starts gRPC server. This is a blocking call.
// To stop serving, cancel passed context.
func (s *Server) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen socket for grpc service: %w", err)
	}

	srv := grpc.NewServer()
	api.RegisterHelloServiceServer(srv, s)
	healthpb.RegisterHealthServer(srv, s.health)
	reflection.Register(srv)
	return s.serve(ctx, srv, lis)
}

func (s *Server) serve(ctx context.Context, srv *grpc.Server, lis net.Listener) error {
	e := make(chan error, 1)
	go func() {
		e <- srv.Serve(lis)
	}()

	s.logger.Info(
		"gRPC server is running",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
	)
	s.health.SetServingStatus(api.HelloService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	select {
	case <-ctx.Done():
		s.health.Shutdown()
		srv.GracefulStop()
		return nil
	case err := <-e:
		return err
	}
}

// Hello ...
func (s *Server) Hello(_ context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	return &api.HelloResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
