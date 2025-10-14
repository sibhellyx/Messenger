package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sibhellyx/Messenger/internal/app"
	"github.com/sibhellyx/Messenger/internal/config"
	"github.com/sibhellyx/Messenger/internal/logger"
)

func main() {
	// init cfg
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// init logger
	logger := logger.NewLogger(cfg.Env.Environment)
	slog.SetDefault(logger)
	// init context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// signal for stopping server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	server := app.NewServer(ctx, cfg, logger)
	go func() {
		<-sigChan
		server.Shutdown()
		cancel()
	}()

	server.Serve()
}
