package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"backend/api"
	"backend/internal/auth"
	auth_persistence "backend/internal/auth/persistence"
	"backend/internal/message"
	message_persistence "backend/internal/message/persistence"
	"backend/internal/moderation"
	moderation_persistence "backend/internal/moderation/persistence"
	"backend/internal/room"
	"backend/internal/shared/application"
	"backend/internal/shared/config"
	"backend/internal/shared/http"
	"backend/internal/shared/infra/db"
	"backend/internal/shared/infra/redis"
	services "backend/internal/shared/infra/services"
	"backend/internal/users"
	users_persistence "backend/internal/users/persistence"
	"backend/internal/websocket"
	"backend/migrations"

	room_app "backend/internal/room/application"
	room_persistence "backend/internal/room/persistence"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	ctx := context.Background()

	// --- Configuration ---
	cfg, err := config.NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	setupLogging(cfg)
	slog.Info("starting server")

	// --- Shared Dependencies ---
	validate := validator.New()

	database, err := db.NewDBPool(cfg.DatabaseURL, db.PoolSettings{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	})
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	defer database.Close()

	if err := migrations.Up(database); err != nil {
		log.Fatalf("could not apply database migrations: %s", err)
	}

	redisClient, err := redis.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("could not initialize redis client: %s", err)
	}

	store := db.NewSQLStore(database)

	jwtService, err := services.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	mailerConfig := application.MailerConfig{
		SMTP_HOST:     cfg.SMTPHost,
		SMTP_PORT:     cfg.SMTPPort,
		SMTP_FROM:     cfg.SMTPFrom,
		SMTP_PASSWORD: cfg.SMTPPassword,
	}
	mailerService, err := services.NewMailerService(mailerConfig)
	if err != nil {
		log.Fatalf("Failed to create Mailer service: %v", err)
	}

	// --- Repositories ---
	authRepository := auth_persistence.NewPgAuthRepository(database)
	usersRepository := users_persistence.NewPgUsersRepository(database)
	tokenRepository := auth_persistence.NewRedisTokenRepository(redisClient)
	messageRepository := message_persistence.NewPgMessageRepository(database)
	reactionRepository := message_persistence.NewPgReactionRepository(database)
	roomRepository := room_persistence.NewPgRoomRepository(database)
	memberRepository := room_persistence.NewPgRoomMemberRepository(database)
	inviteRepository := room_persistence.NewPgInviteRepository(database)
	invitationRepository := room_persistence.NewPgInvitationRepository(database)

	// --- Websocket Hub ---
	wsHub := websocket.NewHub()
	wsUpgrader := websocket.NewUpgrader(cfg.CORSAllowedOrigins)
	// Events are relayed through Redis so several backend instances can serve the same rooms.
	wsHub.UseBus(websocket.NewRedisBus(redisClient))
	go wsHub.Run()
	if err := wsHub.StartBus(ctx); err != nil {
		log.Fatalf("could not start websocket event bus: %s", err)
	}

	// --- Use Cases and Adapters needed across modules ---
	listUserRoomsUseCase := room_app.NewListUserRoomsUseCase(roomRepository)
	roomListerAdapter := room_app.NewRoomListerAdapter(listUserRoomsUseCase)

	// --- Server Setup ---
	server := http.NewServer(cfg)
	apiV1 := server.Router.Group("/api/v1")
	apiV1.Use(http.RateLimit(http.NewRedisCounter(redisClient), []http.Rule{
		{Method: "POST", Path: "/api/v1/auth/login", Limit: 10, Window: time.Minute},
		{Method: "POST", Path: "/api/v1/auth/register", Limit: 20, Window: 10 * time.Minute},
		{Method: "POST", Path: "/api/v1/auth/refresh-token", Limit: 30, Window: time.Minute},
		{Method: "POST", Path: "/api/v1/auth/forgot-password", Limit: 5, Window: 10 * time.Minute},
		{Method: "POST", Path: "/api/v1/auth/reset-password", Limit: 10, Window: 10 * time.Minute},
		{Method: "POST", Path: "/api/v1/auth/send-verify-email/:user_id", Limit: 5, Window: 10 * time.Minute},
		{Method: "POST", Path: "/api/v1/ws-ticket", Limit: 30, Window: time.Minute},
		{Method: "POST", Path: "/api/v1/messages", Limit: 120, Window: time.Minute},
		{Method: "GET", Path: "/api/v1/users/export", Limit: 5, Window: time.Hour},
		{Method: "DELETE", Path: "/api/v1/users/me", Limit: 5, Window: time.Hour},
		{Method: "GET", Path: "/api/v1/messages/search", Limit: 60, Window: time.Minute},
	}))

	// Websocket endpoints: an authenticated call to /ws-ticket returns a short-lived,
	// single-use ticket that /ws redeems, so the JWT never appears in the URL.
	wsTickets := websocket.NewRedisTicketStore(redisClient)
	presence := websocket.NewRedisPresence(redisClient)
	apiV1.POST("/ws-ticket", http.JWTMiddleware(jwtService), websocket.IssueTicket(wsTickets))
	apiV1.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(wsHub, wsUpgrader, wsTickets, roomListerAdapter, presence, c)
	})

	http.RegisterHealthRoutes(server.Router, map[string]http.Pinger{
		"database": http.PingFunc(database.PingContext),
		"redis":    http.PingFunc(func(ctx context.Context) error { return redisClient.Ping(ctx).Err() }),
	})

	// Interactive API reference (Scalar) and the OpenAPI contract it renders.
	api.RegisterDocs(server.Router)

	// --- Module Registration ---
	// Each module is responsible for setting up its own dependencies and routes.
	auth.RegisterModule(authRepository, tokenRepository, apiV1, validate, jwtService, mailerService, store, cfg.FrontendURL, cfg.IsProduction())
	users.RegisterModule(usersRepository, apiV1, validate, mailerService, store, http.JWTMiddleware(jwtService), wsHub, presence)
	moderationRepository := moderation_persistence.NewPgRepository(database)
	room.RegisterModule(roomRepository, memberRepository, inviteRepository, invitationRepository, apiV1, validate, http.JWTMiddleware(jwtService), listUserRoomsUseCase, wsHub, presence, moderationRepository)
	moderation.RegisterModule(moderationRepository, wsHub, apiV1, http.JWTMiddleware(jwtService))
	message.RegisterModule(messageRepository, reactionRepository, wsHub, apiV1, validate, http.JWTMiddleware(jwtService), moderationRepository)

	// --- Start Server ---
	server.Run()
}

// setupLogging sends every log line (including the standard `log` package) through
// slog: JSON in production for log collectors, readable text otherwise.
func setupLogging(cfg *config.Config) {
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, nil)
	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))
}
