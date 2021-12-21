package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/sashayakovtseva/hello-web/internal/config"
	api "github.com/sashayakovtseva/hello-web/pkg/grpc"
)

// Server knows how to serve http requests for geo detection.
type Server struct {
	logger *zap.Logger
	config *config.AppConfig
}

// NewServer returns new Server object that will use passed config
// during setup. To start serving requests call Server.Serve.
func NewServer(
	logger *zap.Logger,
	config *config.AppConfig,
) *Server {
	return &Server{
		logger: logger,
		config: config,
	}
}

// Serve starts HTTP server. This is a blocking call.
// To stop serving, cancel the passed context.
func (s *Server) Serve(ctx context.Context) error {
	gw := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONBuiltin{}))
	opts := []grpc.DialOption{grpc.WithInsecure()} //nolint:staticcheck

	endpoint := ":" + strconv.Itoa(s.config.GRPC.Port)
	if err := api.RegisterHelloServiceHandlerFromEndpoint(ctx, gw, endpoint, opts); err != nil {
		return fmt.Errorf("failed to register gateway service: %w", err)
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.HTTP.Port),
		Handler: gw,
	}

	e := make(chan error, 1)
	go func() {
		e <- srv.ListenAndServe()
	}()

	s.logger.Info(
		"HTTP server is running",
		zap.String("host", s.config.HTTP.Host),
		zap.Int("port", s.config.HTTP.Port),
	)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	case err := <-e:
		return err
	}
}
