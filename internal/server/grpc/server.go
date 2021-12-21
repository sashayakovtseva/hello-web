package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/atomic"
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
	config *config.AppConfig
	health *health.Server

	defaultLastName atomic.String
}

// NewServer returns new Server object that will use passed config
// during setup. To start serving requests call Server.Serve.
func NewServer(
	logger *zap.Logger,
	config *config.AppConfig,
) *Server {
	grpcZap.ReplaceGrpcLoggerV2(logger)

	s := &Server{
		logger: logger,
		config: config,
		health: health.NewServer(),
	}

	s.defaultLastName.Store(s.readDefaultLastName())

	return s
}

// Serve starts gRPC server. This is a blocking call.
// To stop serving, cancel passed context.
func (s *Server) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(s.config.GRPC.Port))
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
		zap.String("host", s.config.GRPC.Host),
		zap.Int("port", s.config.GRPC.Port),
	)
	s.health.SetServingStatus(api.HelloService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-hup:
				s.logger.Info("caught SIGHUP signal, updating config")
				lastName := s.readDefaultLastName()
				s.defaultLastName.Store(lastName)
			}
		}
	}()

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
	name := req.GetName()
	if name == "" {
		name = s.config.DefaultFirstName + " " + s.defaultLastName.String()
	}

	return &api.HelloResponse{
		Greeting: fmt.Sprintf("Hello, %s!", name),
	}, nil
}

type lastNameConfig struct {
	LastName string `json:"last_name"`
}

func (s *Server) readDefaultLastName() string {
	const defaultLastName = "Smith"

	if s.config.DefaultLastNameConfig == "" {
		s.logger.Debug("no config file set, falling back to default")
		return defaultLastName
	}

	f, err := os.Open(s.config.DefaultLastNameConfig)
	if err != nil {
		s.logger.Error("failed to open config file, falling back to default", zap.Error(err))
		return defaultLastName
	}
	defer f.Close() //nolint:errcheck,gosec

	var c lastNameConfig
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		s.logger.Error("failed to decode config file, falling back to default", zap.Error(err))
		return defaultLastName
	}

	s.logger.Info("read config", zap.String("last_name", c.LastName))
	return c.LastName
}
