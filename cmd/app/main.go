package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sibhellyx/Messenger/internal/app"
	"github.com/sibhellyx/Messenger/internal/config"
)

func main() {
	// init cfg
	cfg := config.LoadConfig()
	// init context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// signal for stopping server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	server := app.NewServer(ctx, cfg)
	go func() {
		<-sigChan
		server.Shutdown()
		cancel()
	}()

	server.Serve()

}
