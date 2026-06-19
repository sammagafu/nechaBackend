package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/database"
	"github.com/nechaafrica/backend/internal/handler"
	"github.com/nechaafrica/backend/internal/integration/kkooapp"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/repository"
	"github.com/nechaafrica/backend/internal/router"
	"github.com/nechaafrica/backend/internal/service"
	jwtmanager "github.com/nechaafrica/backend/pkg/jwt"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	if err := database.Seed(db, cfg.SeedDemoUsers()); err != nil {
		log.Fatalf("database seed failed: %v", err)
	}

	var kkClient kkooapp.Client
	if cfg.Kkooapp.APIKey != "" {
		kkClient = kkooapp.NewClient(cfg.Kkooapp)
	} else {
		log.Println("KKOOAPP_API_KEY not set — using mock Kkooapp client")
		kkClient = kkooapp.NewMockClient()
	}

	jwtMgr := jwtmanager.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)

	hotelRepo := repository.NewHotelRepository(db)
	catalogRepo := repository.NewHotelCatalogRepository(db)
	discoveryRepo := repository.NewDiscoveryRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	userRepo := repository.NewUserRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	chatRepo := repository.NewChatRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	guestStayRepo := repository.NewGuestStayRepository(db)

	guestStaySvc := service.NewGuestStayService(guestStayRepo, hotelRepo)
	authSvc := service.NewAuthService(userRepo, jwtMgr, guestStaySvc)
	hotelSvc := service.NewHotelService(hotelRepo, catalogRepo)
	importSvc := service.NewImportService(hotelRepo, catalogRepo)
	discoverySvc := service.NewDiscoveryService(discoveryRepo, hotelRepo)
	notificationSvc := service.NewNotificationService(notificationRepo)
	alertSvc := service.NewAlertService(alertRepo)
	webhookSvc := service.NewWebhookService(webhookRepo, cfg.Webhook.InboundSecret)
	eventSvc := service.NewEventService(notificationSvc, webhookSvc, userRepo)
	chatSvc := service.NewChatService(chatRepo, hotelRepo, userRepo, eventSvc)
	reservationSvc := service.NewReservationService(hotelRepo, reservationRepo, kkClient, eventSvc, guestStaySvc)

	var selcomClient selcom.Client
	switch {
	case cfg.Selcom.MockMode:
		log.Println("Selcom mock payment enabled")
		selcomClient = selcom.NewMockClient(cfg.Selcom.PublicAppURL)
	case cfg.Selcom.APIKey != "" && cfg.Selcom.APISecret != "" && cfg.Selcom.Vendor != "":
		log.Println("Selcom payment gateway enabled")
		selcomClient = selcom.NewClient(cfg.Selcom)
	default:
		log.Println("SELCOM credentials not set — product checkout will skip live payment")
		selcomClient = selcom.NewMockClient(cfg.Selcom.PublicAppURL)
	}

	paymentSvc := service.NewPaymentService(cfg.Selcom, selcomClient, orderRepo, hotelRepo, eventSvc)
	orderSvc := service.NewOrderService(hotelRepo, orderRepo, kkClient, eventSvc, paymentSvc, guestStaySvc)
	adminSvc := service.NewAdminService(hotelRepo, orderRepo, reservationRepo, eventSvc, guestStaySvc)

	app := fiber.New(fiber.Config{
		AppName: "NechaAfrica API",
	})

	router.Setup(app, router.Handlers{
		Auth:           handler.NewAuthHandler(authSvc, cfg.OAuth),
		Hotel:          handler.NewHotelHandler(hotelSvc, guestStaySvc),
		Discovery:      handler.NewDiscoveryHandler(discoverySvc),
		Reservation:    handler.NewReservationHandler(reservationSvc),
		Order:          handler.NewOrderHandler(orderSvc),
		Admin:          handler.NewAdminHandler(adminSvc, authSvc, importSvc),
		Messaging:      handler.NewMessagingHandler(notificationSvc, alertSvc, chatSvc, webhookSvc),
		Payment:        handler.NewPaymentHandler(paymentSvc, cfg.Selcom),
		JWT:            jwtMgr,
		AllowedOrigins: cfg.Server.AllowedOrigin,
		DB:             db,
		Selcom:         cfg.Selcom,
	})

	addr := ":" + cfg.Server.Port
	log.Printf("server starting on %s (env=%s)", addr, cfg.Server.Env)
	if err := app.Listen(addr); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
