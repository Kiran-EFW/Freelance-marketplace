package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/payment"
	"github.com/seva-platform/backend/internal/adapter/push"
	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/adapter/storage"
	"github.com/seva-platform/backend/internal/adapter/whatsapp"
	"github.com/seva-platform/backend/internal/config"
	"github.com/seva-platform/backend/internal/handler"
	"github.com/seva-platform/backend/internal/middleware"
	"github.com/seva-platform/backend/internal/repository/postgres"
	rediscache "github.com/seva-platform/backend/internal/repository/redis"
	svcadapter "github.com/seva-platform/backend/internal/service/adapter"
	adminsvc "github.com/seva-platform/backend/internal/service/admin"
	smssession "github.com/seva-platform/backend/internal/service/sms_session"
	aisvc "github.com/seva-platform/backend/internal/service/ai"
	cropsvc "github.com/seva-platform/backend/internal/service/crop"
	disputesvc "github.com/seva-platform/backend/internal/service/dispute"
	gamificationsvc "github.com/seva-platform/backend/internal/service/gamification"
	jobsvc "github.com/seva-platform/backend/internal/service/job"
	jurisdictionsvc "github.com/seva-platform/backend/internal/service/jurisdiction"
	messagingsvc "github.com/seva-platform/backend/internal/service/messaging"
	notifsvc "github.com/seva-platform/backend/internal/service/notification"
	paymentsvc "github.com/seva-platform/backend/internal/service/payment"
	reviewsvc "github.com/seva-platform/backend/internal/service/review"
	routingsvc "github.com/seva-platform/backend/internal/service/routing"
	searchsvc "github.com/seva-platform/backend/internal/service/search"
	subscriptionsvc "github.com/seva-platform/backend/internal/service/subscription"
	usersvc "github.com/seva-platform/backend/internal/service/user"
	"github.com/seva-platform/backend/internal/worker"
	"github.com/seva-platform/backend/pkg/logger"
)

func main() {
	// ---- Configuration ----
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// ---- Logger ----
	logger.Setup(cfg.Environment)
	log.Info().Str("env", cfg.Environment).Msg("starting Seva API")

	// ---- PostgreSQL ----
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create database pool")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}
	log.Info().Msg("connected to PostgreSQL")

	// ---- Redis ----
	redisOpts, err := parseRedisURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse Redis URL")
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping Redis")
	}
	log.Info().Msg("connected to Redis")

	// ---- SMS Adapter ----
	smsProvider, err := sms.NewSMSProvider(cfg)
	if err != nil {
		log.Warn().Err(err).Msg("SMS provider not configured, falling back to noop")
		smsProvider = &sms.NoopProvider{}
	}

	// ---- WhatsApp Adapter ----
	whatsappProvider := whatsapp.NewProvider(cfg)
	_ = whatsappProvider // used by WhatsApp webhook handler below

	// ---- FCM Push Adapter ----
	pushProvider := push.NewPushProvider(cfg.FCMProjectID, cfg.GetFCMServiceAccountKey())

	// ---- Repositories ----
	queries := postgres.New(dbPool)
	cacheStore := rediscache.NewCacheStore(rdb)

	// Repository adapters bridge domain interfaces to sqlc-generated Queries.
	userRepo := postgres.NewUserRepository(queries)
	providerRepo := postgres.NewProviderRepository(queries)
	jobRepo := postgres.NewJobRepository(queries)
	reviewRepo := postgres.NewReviewRepository(queries)
	transactionRepo := postgres.NewTransactionRepository(queries)
	notificationRepo := postgres.NewNotificationRepository(queries, dbPool)
	disputeRepo := postgres.NewDisputeRepository(queries)
	gamificationRepo := postgres.NewGamificationRepository(queries)
	categoryRepo := postgres.NewCategoryRepository(queries)
	routeRepo := postgres.NewRouteRepository(queries, dbPool)
	cacheAdapter := postgres.NewCacheStoreAdapter(cacheStore)

	// ---- Object Storage ----
	storageClient := storage.NewR2Storage(
		cfg.R2AccountID,
		cfg.R2AccessKeyID,
		cfg.R2AccessKeySecret,
		cfg.R2Bucket,
		cfg.R2PublicURL,
	)

	// ---- Payment Gateway ----
	paymentGateway := payment.NewPaymentGateway(cfg)

	// ---- Crop Repository ----
	cropRepo := postgres.NewCropRepository(dbPool)

	// ---- Services ----
	userSvc := usersvc.NewUserService(userRepo, cacheAdapter)
	jobSvc := jobsvc.NewJobService(jobRepo)
	reviewSvc := reviewsvc.NewReviewService(reviewRepo, jobRepo, disputeRepo, providerRepo)
	paymentSvc := paymentsvc.NewPaymentService(transactionRepo, jobRepo, paymentGateway)
	notifSvc := notifsvc.NewNotificationService(notificationRepo, userRepo, smsProvider, pushProvider, queries)
	disputeSvc := disputesvc.NewDisputeService(disputeRepo, jobRepo)
	gamificationSvc := gamificationsvc.NewGamificationService(gamificationRepo, providerRepo, reviewRepo)
	routingSvc := routingsvc.NewRoutingService(routeRepo, providerRepo)
	cropSvc := cropsvc.NewCropService(cropRepo)
	searchSvc := searchsvc.NewSearchService(queries, dbPool)
	adminSvc := adminsvc.NewAdminService(queries, dbPool, disputeRepo)
	aiSvc := aisvc.NewAIService(queries, dbPool)
	messagingSvc := messagingsvc.NewMessagingService(dbPool)
	subscriptionSvc := subscriptionsvc.NewSubscriptionService(dbPool)

	// Suppress unused warnings for repos used only via adapters below.
	_ = categoryRepo

	// ---- Service-to-Handler Adapters ----
	// These adapters bridge the handler-level interfaces (handler-local types)
	// to the service-level implementations (domain types).
	userSvcAdapter := svcadapter.NewUserServiceAdapter(userSvc)
	jobSvcAdapter := svcadapter.NewJobServiceAdapter(jobSvc, queries)
	reviewSvcAdapter := svcadapter.NewReviewServiceAdapter(reviewSvc, reviewRepo)
	paymentSvcAdapter := svcadapter.NewPaymentServiceAdapter(paymentSvc, transactionRepo)
	notifSvcAdapter := svcadapter.NewNotificationServiceAdapter(notifSvc, notificationRepo)
	disputeSvcAdapter := svcadapter.NewDisputeServiceAdapter(disputeSvc, disputeRepo)
	gamificationSvcAdapter := svcadapter.NewGamificationServiceAdapter(gamificationSvc, gamificationRepo)
	routeSvcAdapter := svcadapter.NewRouteServiceAdapter(routingSvc, routeRepo)
	providerSvcAdapter := svcadapter.NewProviderServiceAdapter(providerRepo, reviewRepo, transactionRepo, jobRepo)
	cropSvcAdapter := svcadapter.NewCropServiceAdapter(cropSvc)
	searchSvcAdapter := svcadapter.NewSearchServiceAdapter(searchSvc)
	adminSvcAdapter := svcadapter.NewAdminServiceAdapter(adminSvc)
	aiSvcAdapter := svcadapter.NewAIServiceAdapter(aiSvc)
	messagingSvcAdapter := svcadapter.NewMessageServiceAdapter(messagingSvc)
	subscriptionSvcAdapter := svcadapter.NewSubscriptionServiceAdapter(subscriptionSvc)
	escrowSvcAdapter := svcadapter.NewEscrowServiceAdapter(queries)
	recurringSvcAdapter := svcadapter.NewRecurringServiceAdapter(queries)
	analyticsSvcAdapter := svcadapter.NewAnalyticsServiceAdapter(queries)

	// ---- Handlers ----
	healthHandler := handler.NewHealthHandler(dbPool, rdb)
	authHandler := handler.NewAuthHandler(cfg, queries, cacheStore, smsProvider)
	userHandler := handler.NewUserHandler(userSvcAdapter)
	jobHandler := handler.NewJobHandler(jobSvcAdapter)
	providerHandler := handler.NewProviderHandler(providerSvcAdapter, storageClient)
	reviewHandler := handler.NewReviewHandler(reviewSvcAdapter)
	paymentHandler := handler.NewPaymentHandler(paymentSvcAdapter)
	notificationHandler := handler.NewNotificationHandler(notifSvcAdapter)
	disputeHandler := handler.NewDisputeHandler(disputeSvcAdapter, storageClient)
	gamificationHandler := handler.NewGamificationHandler(gamificationSvcAdapter)
	routeHandler := handler.NewRouteHandler(routeSvcAdapter)
	cropHandler := handler.NewCropHandler(cropSvcAdapter)
	escrowHandler := handler.NewEscrowHandler(escrowSvcAdapter)
	recurringHandler := handler.NewRecurringHandler(recurringSvcAdapter)

	// Previously nil-guarded handlers now wired to real services.
	searchHandler := handler.NewSearchHandler(searchSvcAdapter)
	adminHandler := handler.NewAdminHandler(adminSvcAdapter)
	aiHandler := handler.NewAIHandler(aiSvcAdapter)
	// WebSocket hub and handler
	wsHub := handler.NewHub()
	wsHandler := handler.NewWebSocketHandler(wsHub, cfg)
	messageHandler := handler.NewMessageHandler(messagingSvcAdapter)
	messageHandler.SetHub(wsHub) // Wire hub into message handler for real-time push
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionSvcAdapter)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvcAdapter)

	// ---- Jurisdiction Service ----
	jurisdictionService := jurisdictionsvc.NewJurisdictionService(queries, cacheStore)

	jurisdictionHandler := handler.NewJurisdictionHandler(
		func(ctx context.Context) (interface{}, error) {
			return jurisdictionService.ListActive(ctx)
		},
		func(ctx context.Context, id string) (interface{}, error) {
			return jurisdictionService.GetJurisdiction(ctx, id)
		},
		func(ctx context.Context, id string) ([]string, error) {
			return jurisdictionService.GetEnabledCategories(ctx, id)
		},
		jurisdictionService.DetectJurisdiction,
	)

	// ---- Organization Handler (B2B Dashboard) ----
	organizationHandler := handler.NewOrganizationHandler(queries)

	// ---- Safety Handler ----
	safetyHandler := handler.NewSafetyHandler(queries, smsProvider)

	// ---- Device Token Handler ----
	deviceTokenHandler := handler.NewDeviceTokenHandler(queries)

	// ---- Asynq Worker ----
	workerDeps := &worker.Deps{
		SMSProvider:    smsProvider,
		DB:             dbPool,
		Redis:          rdb,
		PaymentGateway: paymentGateway,
		Email: &worker.EmailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		},
		Cfg: cfg,
	}
	w := worker.NewWorker(redisOpts.Addr, workerDeps)
	go func() {
		if err := w.Start(); err != nil {
			log.Error().Err(err).Msg("asynq worker stopped with error")
		}
	}()

	// ---- Periodic Task: Process Recurring Jobs (every 15 minutes) ----
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisOpts.Addr})
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			task, err := worker.NewProcessRecurringJobsTask()
			if err != nil {
				log.Error().Err(err).Msg("failed to create recurring jobs task")
				continue
			}
			if _, err := asynqClient.Enqueue(task); err != nil {
				log.Error().Err(err).Msg("failed to enqueue recurring jobs task")
			}
		}
	}()

	// ---- Fiber App ----
	app := fiber.New(fiber.Config{
		AppName:               "Seva API",
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             15 * 1024 * 1024, // 15MB for file uploads
		DisableStartupMessage: cfg.IsProd(),
		ErrorHandler:          customErrorHandler,
	})

	// ---- Global Middleware ----
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.FiberLogger())
	app.Use(middleware.NewCORS(cfg))
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax(),
		Expiration: 1 * time.Second,
	}))

	// ---- Route Groups ----
	api := app.Group("/api/v1")

	// --- Public Routes ---

	// Health (public)
	healthHandler.RegisterRoutes(api.Group("/health"))

	// Auth (public)
	authHandler.RegisterRoutes(api.Group("/auth"))

	// Search (public)
	searchHandler.RegisterSearchRoutes(api.Group("/search"))

	// Categories (public)
	searchHandler.RegisterCategoryRoutes(api.Group("/categories"))

	// Jurisdictions (public)
	jurisdictionHandler.RegisterRoutes(api.Group("/jurisdictions"))

	// --- Authenticated Routes ---
	jwtMiddleware := middleware.NewJWTAuth(cfg)

	// Users
	userHandler.RegisterRoutes(api.Group("/users", jwtMiddleware))

	// Jobs
	jobHandler.RegisterRoutes(api.Group("/jobs", jwtMiddleware))

	// Providers
	providerHandler.RegisterRoutes(api.Group("/providers", jwtMiddleware))

	// Reviews
	reviewHandler.RegisterRoutes(api.Group("/reviews", jwtMiddleware))
	// Provider review and rating sub-routes are registered on the providers group.
	reviewHandler.RegisterProviderReviewRoutes(api.Group("/providers", jwtMiddleware))

	// Payments
	paymentHandler.RegisterRoutes(api.Group("/payments", jwtMiddleware))

	// Disputes
	disputeHandler.RegisterRoutes(api.Group("/disputes", jwtMiddleware))

	// Points / Gamification
	gamificationHandler.RegisterRoutes(api.Group("/points", jwtMiddleware))

	// Routes (service worker route management)
	routeHandler.RegisterRoutes(api.Group("/routes", jwtMiddleware))

	// Notifications
	notificationHandler.RegisterRoutes(api.Group("/notifications", jwtMiddleware))

	// Device tokens (push notifications)
	deviceTokenHandler.RegisterRoutes(api.Group("/notifications", jwtMiddleware))

	// AI endpoints (authenticated)
	aiHandler.RegisterRoutes(api.Group("/ai", jwtMiddleware))

	// Messages (authenticated)
	messageHandler.RegisterRoutes(api.Group("/messages", jwtMiddleware))

	// Subscriptions (authenticated)
	subscriptionHandler.RegisterRoutes(api.Group("/subscriptions", jwtMiddleware))

	// Escrow (authenticated)
	escrowHandler.RegisterRoutes(api.Group("/escrow", jwtMiddleware))

	// Recurring schedules (authenticated)
	recurringHandler.RegisterRoutes(api.Group("/recurring", jwtMiddleware))

	// Organizations (B2B Dashboard)
	organizationHandler.RegisterRoutes(api.Group("/organizations", jwtMiddleware))

	// Safety (SOS, live tracking, emergency contacts)
	safetyHandler.RegisterRoutes(api.Group("/safety", jwtMiddleware))

	// Analytics (authenticated — provider analytics dashboard)
	analyticsHandler.RegisterRoutes(api.Group("/analytics", jwtMiddleware))

	// WebSocket (JWT validated via query param)
	app.Use("/ws", wsHandler.UpgradeMiddleware())
	app.Get("/ws", wsHandler.HandleWebSocket())

	// Crop Calendar (public — no auth required for browsing)
	cropHandler.RegisterRoutes(api.Group("/crops"))

	// --- Admin Routes (admin role required) ---
	adminGroup := api.Group("/admin", jwtMiddleware, middleware.RequireAdmin())
	adminHandler.RegisterRoutes(adminGroup)
	disputeHandler.RegisterAdminRoutes(adminGroup.Group("/disputes"))

	// --- Webhook Routes (signature-verified, no JWT) ---
	webhookGroup := app.Group("/webhooks")
	paymentHandler.RegisterWebhookRoutes(webhookGroup)

	// WhatsApp webhook (verification + incoming messages)
	whatsappWebhookHandler := handler.NewWhatsAppWebhookHandler(cfg, &handler.DefaultWhatsAppMessageHandler{})
	whatsappWebhookHandler.RegisterRoutes(webhookGroup)

	// Subscription payment webhook
	subscriptionHandler.RegisterWebhookRoutes(webhookGroup)

	// SMS interface webhook (basic phone users)
	smsSessionMgr := smssession.NewSMSSessionManager(rdb)
	smsInterfaceHandler := handler.NewSMSInterfaceHandler(cfg, smsSessionMgr)
	smsInterfaceHandler.RegisterRoutes(webhookGroup.Group("/sms"))

	// IVR handler (voice calls for basic phone users)
	ivrHandler := handler.NewIVRHandler(cfg)
	ivrHandler.RegisterRoutes(webhookGroup.Group("/ivr"))

	// ---- Graceful Shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	<-quit
	log.Info().Msg("shutting down gracefully...")

	w.Shutdown()

	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Error().Err(err).Msg("error during server shutdown")
	}

	log.Info().Msg("server stopped")
}

// parseRedisURL converts a redis:// URL into go-redis options.
func parseRedisURL(rawURL string) (*redis.Options, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	addr := u.Host
	if u.Port() == "" {
		addr = u.Host + ":6379"
	}

	password, _ := u.User.Password()

	db := 0
	if len(u.Path) > 1 {
		fmt.Sscanf(u.Path[1:], "%d", &db)
	}

	return &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}, nil
}

// customErrorHandler returns consistent JSON error responses.
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("request error")

	return c.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}
