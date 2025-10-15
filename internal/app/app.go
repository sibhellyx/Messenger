package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sibhellyx/Messenger/api"
	"github.com/sibhellyx/Messenger/internal/config"
	"github.com/sibhellyx/Messenger/internal/db/authrepo"
	"github.com/sibhellyx/Messenger/internal/db/chatrepo"
	"github.com/sibhellyx/Messenger/internal/db/migrate"
	"github.com/sibhellyx/Messenger/internal/db/msgrepo"
	"github.com/sibhellyx/Messenger/internal/kafka"
	authservice "github.com/sibhellyx/Messenger/internal/services/authService"
	chatservice "github.com/sibhellyx/Messenger/internal/services/chatService"
	messageservice "github.com/sibhellyx/Messenger/internal/services/messageService"
	wsservice "github.com/sibhellyx/Messenger/internal/services/wsService"
	authhandler "github.com/sibhellyx/Messenger/internal/transport/authHandler"
	chathandler "github.com/sibhellyx/Messenger/internal/transport/chatHandler"
	messagehandler "github.com/sibhellyx/Messenger/internal/transport/messageHandler"
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
	slog.Info("starting server", "port", srv.cfg.Srv.Port)

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
	err = migrate.Migrate(srv.db)
	if err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	// init and start hub for chat
	slog.Debug("connect and start hub")
	hub := ws.NewHub(srv.cfg.Ws)
	go hub.Run()

	//init hasher and manager
	slog.Debug("init hasher for passwords")
	hasher := hash.NewHasher(srv.cfg.Auth.Salt)
	slog.Debug("init manager for auth")
	manager := auth.NewManager(srv.cfg.Auth.SigningKey)
	// init kafka
	slog.Debug("init kafka producer")
	producer := kafka.NewProducer(srv.cfg.Kafka)
	defer producer.Close() //add closing producer

	// init repos for auth
	slog.Debug("connecting to auth repository")
	authRepository := authrepo.NewAuthRepository(srv.db)
	slog.Debug("connecting to chat repository")
	chatRepository := chatrepo.NewChatRepository(srv.db)
	slog.Debug("connecting to message repository")
	messageRepository := msgrepo.NewMessageRepository(srv.db)

	// init service for auth
	slog.Debug("connecting to auth service")
	authService := authservice.NewAuthService(
		authRepository,
		hasher,
		manager,
		time.Duration(srv.cfg.Jwt.AccessTTL*int(time.Minute)),
		time.Duration(srv.cfg.Jwt.RefreshTTL*int(time.Hour*24)),
		srv.cfg.Jwt.ActiveSessions,
	)
	slog.Debug("connecting to chat service")
	chatService := chatservice.NewChatService(chatRepository)
	slog.Debug("connecting to ws service")
	wsService := wsservice.NewWsService(hub)
	slog.Debug("connecting to message service")
	messageService := messageservice.NewMessageService(wsService, producer, messageRepository)

	slog.Debug("init kafka consumer")
	consumer := kafka.NewConsumer(srv.cfg.Kafka, messageService)

	slog.Debug("set consumer to message service")
	messageService.SetConsumer(consumer)

	// start consumer
	slog.Debug("start consumer in message service")
	go messageService.StartConsumer(context.Background())
	defer messageService.StopConsumer()

	// init Handlers
	slog.Debug("connecting to auth handler")
	authHandler := authhandler.NewAuthHandler(authService)
	slog.Debug("connecting to chat handler")
	chatHandler := chathandler.NewChatHandler(chatService)
	slog.Debug("connecting to ws handler")
	wsHandler := wshandler.NewWsHandler(wsService)
	slog.Debug("connecting to message handler")
	messageHandler := messagehandler.NewMessageHandler(messageService)

	//init routes for messanger
	slog.Debug("creating routes")
	routes := api.CreateRoutes(authHandler, chatHandler, wsHandler, messageHandler, manager, authRepository)

	// create http server
	slog.Debug("init server")
	srv.srv = &http.Server{
		Addr:    ":" + srv.cfg.Srv.Port,
		Handler: routes,
	}

	// start server
	slog.Info("starting HTTP server", "port", srv.cfg.Srv.Port)
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
