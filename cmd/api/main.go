package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/database"
	"github.com/nechaafrica/backend/internal/handler"
	"github.com/nechaafrica/backend/internal/integration/email"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/integration/sms"
	"github.com/nechaafrica/backend/internal/integration/whatsapp"
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
	inquiryRepo := repository.NewInquiryRepository(db)
	guestRequestRepo := repository.NewGuestRequestRepository(db)
	deliveryZoneRepo := repository.NewDeliveryZoneRepository(db)
	platformConfigRepo := repository.NewPlatformConfigRepository(db)

	guestStayRepo := repository.NewGuestStayRepository(db)

	influencerRepo := repository.NewInfluencerRepository(db)
	bookingReferralRepo := repository.NewBookingReferralRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	catalogueRepo := repository.NewCatalogueRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	payoutRepo := repository.NewPayoutRepository(db)
	eventLogRepo := repository.NewEventLogRepository(db)
	rewardRepo := repository.NewRewardRepository(db)

	if err := service.SeedDefaultCommissionRules(commissionRepo); err != nil {
		log.Printf("commission rule seed warning: %v", err)
	}

	emailClient := email.NewClient(email.Config{
		Enabled:  cfg.Email.Enabled,
		Host:     cfg.Email.Host,
		Port:     cfg.Email.Port,
		Username: cfg.Email.Username,
		Password: cfg.Email.Password,
		From:     cfg.Email.From,
		AdminTo:  cfg.Email.AdminTo,
	})

	smsClient := sms.NewClient(sms.Config{
		Enabled: cfg.SMS.Enabled,
		APIURL:  cfg.SMS.APIURL,
		APIKey:  cfg.SMS.APIKey,
		Sender:  cfg.SMS.Sender,
		AdminTo: cfg.SMS.AdminTo,
	})

	whatsappClient := whatsapp.NewClient(whatsapp.Config{
		Enabled: cfg.WhatsApp.Enabled,
		APIURL:  cfg.WhatsApp.APIURL,
		APIKey:  cfg.WhatsApp.APIKey,
		Sender:  cfg.WhatsApp.Sender,
		AdminTo: cfg.WhatsApp.AdminTo,
	})

	guestStaySvc := service.NewGuestStayService(guestStayRepo, hotelRepo)
	authSvc := service.NewAuthService(userRepo, jwtMgr, guestStaySvc)
	hotelSvc := service.NewHotelService(hotelRepo, catalogRepo)
	importSvc := service.NewImportService(hotelRepo, catalogRepo)
	discoverySvc := service.NewDiscoveryService(discoveryRepo, hotelRepo)
	notificationSvc := service.NewNotificationService(notificationRepo)
	alertSvc := service.NewAlertService(alertRepo)
	webhookSvc := service.NewWebhookService(webhookRepo, cfg.Webhook.InboundSecret)
	eventSvc := service.NewEventService(notificationSvc, webhookSvc, userRepo, emailClient, smsClient, whatsappClient)
	commissionSvc := service.NewCommissionService(commissionRepo, influencerRepo, eventLogRepo)
	eventSvc.SetCommissionService(commissionSvc)
	eventSvc.SetRewardRepository(rewardRepo)
	platformSvc := service.NewPlatformService(deliveryZoneRepo, platformConfigRepo)
	eventSvc.SetPlatform(platformSvc)
	commerceSvc := service.NewCommerceService(influencerRepo, bookingReferralRepo, supplierRepo, catalogueRepo, commissionRepo, rewardRepo, hotelRepo, eventLogRepo)
	chatSvc := service.NewChatService(chatRepo, hotelRepo, userRepo, eventSvc)

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

	payoutSvc := service.NewPayoutService(commissionRepo, payoutRepo, eventLogRepo, hotelRepo, influencerRepo, selcomClient)
	reservationSvc := service.NewReservationService(hotelRepo, reservationRepo, eventSvc, guestStaySvc)

	paymentSvc := service.NewPaymentService(cfg.Selcom, selcomClient, orderRepo, hotelRepo, discoveryRepo, eventSvc)
	orderSvc := service.NewOrderService(hotelRepo, catalogRepo, orderRepo, discoveryRepo, eventSvc, paymentSvc, guestStaySvc, platformSvc, influencerRepo, bookingReferralRepo, rewardRepo)
	adminSvc := service.NewAdminService(hotelRepo, catalogRepo, orderRepo, reservationRepo, eventSvc, guestStaySvc)
	partnerSvc := service.NewPartnerService(userRepo, adminSvc, commissionRepo, guestStayRepo, hotelRepo, bookingReferralRepo)
	supplierPortalSvc := service.NewSupplierPortalService(userRepo, supplierRepo, hotelRepo)
	inquirySvc := service.NewInquiryService(inquiryRepo, eventSvc, platformSvc)
	guestRequestSvc := service.NewGuestRequestService(guestRequestRepo, hotelRepo, eventSvc)

	// Scheduled background jobs: founding-tier auto-transition + inventory reservation sweep.
	maintenanceSvc := service.NewMaintenanceService(hotelRepo, platformConfigRepo, eventLogRepo, db)
	maintenanceSvc.Start()

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
		Inquiry:        handler.NewInquiryHandler(inquirySvc),
		Platform:       handler.NewPlatformHandler(platformSvc, guestRequestSvc),
		Payment:        handler.NewPaymentHandler(paymentSvc, cfg.Selcom, commissionSvc),
		Commerce:       handler.NewCommerceHandler(commerceSvc, payoutSvc, platformSvc),
		Partner:        handler.NewPartnerHandler(partnerSvc, platformSvc),
		Supplier:       handler.NewSupplierPortalHandler(supplierPortalSvc),
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
