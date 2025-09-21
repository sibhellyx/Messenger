package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sibhellyx/Messenger/api"
	"github.com/sibhellyx/Messenger/internal/config"
	"github.com/sibhellyx/Messenger/internal/db"
	authservice "github.com/sibhellyx/Messenger/internal/services/authService"
	authhandler "github.com/sibhellyx/Messenger/internal/transport/authHandler"
	"github.com/sibhellyx/Messenger/pkg/auth"
	"github.com/sibhellyx/Messenger/pkg/hash"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	ctx    context.Context
	cfg    config.Config
	srv    *http.Server
	db     *gorm.DB
	logger *slog.Logger
}

func NewServer(ctx context.Context, cfg config.Config) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	srv := &Server{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}

	return srv
}

func (srv *Server) Serve() {
	srv.logger.Info("starting server", "port", srv.cfg.Port)

	srv.logger.Debug("connecting to database")
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  srv.cfg.GetDbString(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		srv.logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	srv.db = database

	err = db.Migrate(srv.db, srv.logger)
	if err != nil {
		srv.logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	hasher := hash.NewHasher("salt")
	manager := auth.NewManager("some-auth-manager", srv.logger)

	srv.logger.Debug("connecting to auth repository")
	repository := db.NewRepository(srv.db, srv.logger)

	srv.logger.Debug("connecting to auth service")
	authService := authservice.NewAuthService(
		repository,
		srv.logger,
		hasher,
		manager,
		time.Duration(srv.cfg.AccessTTL*int(time.Minute)),
		time.Duration(srv.cfg.RefreshTTL*int(time.Hour*24)),
		srv.cfg.ActiveSessions,
	)

	srv.logger.Debug("connecting to auth handler")
	authHandler := authhandler.NewAuthHandler(authService)

	srv.logger.Debug("creating routes")
	routes := api.CreateRoutes(authHandler, srv.logger, manager, repository)

	srv.logger.Debug("init server")
	srv.srv = &http.Server{
		Addr:    ":" + srv.cfg.Port,
		Handler: routes,
	}

	srv.logger.Info("starting HTTP server", "port", srv.cfg.Port)
	if err := srv.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		srv.logger.Error("HTTP server error - shutting down", "error", err)
		os.Exit(1)
	}
	srv.logger.Info("HTTP server stopped")
}

func (srv *Server) Shutdown() {
	slog.Info("server stopping...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.srv.Shutdown(ctxShutdown)
	if err != nil {
		srv.logger.Error("HTTP server shutdown error", "error", err)
		os.Exit(1)
	}

}
