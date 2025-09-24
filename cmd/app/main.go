package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sibhellyx/Messenger/internal/app"
	"github.com/sibhellyx/Messenger/internal/config"
	"github.com/sibhellyx/Messenger/internal/logger"
)

func main() {
	// load enviroment for logger
	envConfig, err := config.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}
	// init logger
	logger := logger.NewLogger(envConfig.Environment)
	slog.SetDefault(logger)

	// init cfg
	cfg := config.LoadConfig()
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
