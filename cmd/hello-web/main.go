package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"

	"github.com/sashayakovtseva/hello-web/internal/config"
	"github.com/sashayakovtseva/hello-web/internal/server/grpc"
	"github.com/sashayakovtseva/hello-web/internal/server/http"
)

func main() {
	appConfig, err := config.New()
	if err != nil {
		if errors.Is(err, config.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("failed to read app config: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	defer cancel()

	grpcServer := grpc.NewServer(logger, appConfig)

	gr, appctx := errgroup.WithContext(ctx)
	gr.Go(func() error {
		return grpcServer.Serve(appctx)
	})
	gr.Go(func() error {
		httpServer := http.NewServer(logger, appConfig)
		return httpServer.Serve(appctx)
	})
	if err := gr.Wait(); err != nil {
		logger.Error("application exited with error", zap.Error(err))
	}

	logger.Info("shutting down logger")
	logger.Sync() //nolint:errcheck,gosec
}
