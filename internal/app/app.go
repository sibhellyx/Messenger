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
	"github.com/sibhellyx/Messenger/internal/db/authrepo"
	authservice "github.com/sibhellyx/Messenger/internal/services/authService"
	authhandler "github.com/sibhellyx/Messenger/internal/transport/authHandler"
	wshandler "github.com/sibhellyx/Messenger/internal/transport/wsHandler"
	"github.com/sibhellyx/Messenger/internal/ws"
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

	// load config for web sockets
	srv.logger.Debug("loading configs for websockets")
	wsConfs := config.LoadWsConfig()

	// start database
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

	// migration database
	err = db.Migrate(srv.db, srv.logger)
	if err != nil {
		srv.logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	// start hub for chat
	hub := ws.NewHub(srv.logger, wsConfs)
	go hub.Run()

	//init hasher and manager
	hasher := hash.NewHasher("salt")
	manager := auth.NewManager("some-auth-manager", srv.logger)

	// init repos for auth
	srv.logger.Debug("connecting to auth repository")
	repository := authrepo.NewRepository(srv.db, srv.logger)

	// init service for auth
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

	// init authHandler
	srv.logger.Debug("connecting to auth handler")
	authHandler := authhandler.NewAuthHandler(authService)

	// init ws handler
	srv.logger.Debug("connecting to ws handler")
	wsHandler := wshandler.NewWsHandler(hub, srv.logger)

	//init routes for messanger
	srv.logger.Debug("creating routes")
	routes := api.CreateRoutes(authHandler, wsHandler, srv.logger, manager, repository)

	// create http server
	srv.logger.Debug("init server")
	srv.srv = &http.Server{
		Addr:    ":" + srv.cfg.Port,
		Handler: routes,
	}

	// start server
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
