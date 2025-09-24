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

func NewServer(ctx context.Context, cfg config.Config, logger *slog.Logger) *Server {
	srv := &Server{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}

	return srv
}

func (srv *Server) Serve() {
	slog.Info("starting server", "port", srv.cfg.Port)

	// load config for web sockets
	slog.Debug("loading configs for websockets")
	wsConfs := config.LoadWsConfig()

	// start database
	slog.Debug("connecting to database")
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  srv.cfg.GetDbString(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	srv.db = database

	// migration database
	slog.Debug("do migration to database")
	err = db.Migrate(srv.db)
	if err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	// init and start hub for chat
	slog.Debug("connect and start hub")
	hub := ws.NewHub(wsConfs)
	go hub.Run()

	//init hasher and manager
	slog.Debug("init hasher for passwords")
	hasher := hash.NewHasher("salt")
	slog.Debug("init manager for auth")
	manager := auth.NewManager("some-auth-manager")

	// init repos for auth
	slog.Debug("connecting to auth repository")
	repository := authrepo.NewRepository(srv.db)

	// init service for auth
	slog.Debug("connecting to auth service")
	authService := authservice.NewAuthService(
		repository,
		hasher,
		manager,
		time.Duration(srv.cfg.AccessTTL*int(time.Minute)),
		time.Duration(srv.cfg.RefreshTTL*int(time.Hour*24)),
		srv.cfg.ActiveSessions,
	)

	// init authHandler
	slog.Debug("connecting to auth handler")
	authHandler := authhandler.NewAuthHandler(authService)

	// init ws handler
	slog.Debug("connecting to ws handler")
	wsHandler := wshandler.NewWsHandler(hub)

	//init routes for messanger
	slog.Debug("creating routes")
	routes := api.CreateRoutes(authHandler, wsHandler, manager, repository)

	// create http server
	slog.Debug("init server")
	srv.srv = &http.Server{
		Addr:    ":" + srv.cfg.Port,
		Handler: routes,
	}

	// start server
	slog.Info("starting HTTP server", "port", srv.cfg.Port)
	if err := srv.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error - shutting down", "error", err)
		os.Exit(1)
	}
	slog.Info("HTTP server stopped")
}

func (srv *Server) Shutdown() {
	slog.Info("server stopping...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.srv.Shutdown(ctxShutdown)
	if err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
		os.Exit(1)
	}

}
