package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/handler"
	"github.com/nechaafrica/backend/internal/middleware"
	jwtmanager "github.com/nechaafrica/backend/pkg/jwt"
	"gorm.io/gorm"
)

type Handlers struct {
	Auth           *handler.AuthHandler
	Hotel          *handler.HotelHandler
	Discovery      *handler.DiscoveryHandler
	Reservation    *handler.ReservationHandler
	Order          *handler.OrderHandler
	Admin          *handler.AdminHandler
	Messaging      *handler.MessagingHandler
	Inquiry        *handler.InquiryHandler
	Platform       *handler.PlatformHandler
	Payment        *handler.PaymentHandler
	Commerce       *handler.CommerceHandler
	Partner        *handler.PartnerHandler
	JWT            *jwtmanager.Manager
	AllowedOrigins string
	DB             *gorm.DB
	Selcom         config.SelcomConfig
}

func Setup(app *fiber.App, h Handlers) {
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.RequestID())
	app.Use(limiter.New(limiter.Config{
		Max:        120,
		Expiration: time.Minute,
	}))

	origins := h.AllowedOrigins
	if origins == "" {
		origins = "https://necha.africa"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins + ",http://localhost:5173,http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID, X-Guest-Token, X-Webhook-Secret",
		AllowCredentials: true,
	}))

	app.Get("/health", handler.Health)
	app.Get("/health/ready", handler.HealthReady(h.DB))
	app.Get("/api/health", handler.Health)

	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Use(limiter.New(limiter.Config{Max: 20, Expiration: time.Minute}))
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/google", h.Auth.GoogleLogin)
	auth.Post("/apple", h.Auth.AppleLogin)
	auth.Get("/me", middleware.Auth(h.JWT), h.Auth.Me)

	api.Get("/partners/landing", h.Hotel.PartnersLanding)

	hotels := api.Group("/hotels")
	hotels.Get("/slug/:slug", h.Hotel.GetBySlug)
	hotels.Get("/slug/:slug/products", h.Hotel.ListProducts)
	hotels.Get("/slug/:slug/products/:productSlug", h.Hotel.GetProduct)
	hotels.Get("/slug/:slug/products/:productSlug/reviews", h.Hotel.ListProductReviews)
	hotels.Post("/slug/:slug/products/:productSlug/reviews", h.Hotel.CreateProductReview)
	hotels.Get("/slug/:slug/rooms", h.Hotel.ListRooms)
	hotels.Get("/slug/:slug/menu", h.Hotel.ListMenu)
	hotels.Post("/slug/:slug/requests", h.Platform.SubmitGuestRequest)
	hotels.Post("/slug/:slug/scan", middleware.OptionalAuth(h.JWT), h.Hotel.RecordScan)
	hotels.Get("/slug/:slug/discovery", h.Discovery.PortalBySlug)
	hotels.Get("/:code/products", h.Hotel.ListProductsByCode)
	hotels.Get("/:code", h.Hotel.GetByCode)

	discovery := api.Group("/discovery")
	discovery.Get("/items/:slug", h.Discovery.PublicGetBySlug)
	discovery.Post("/events/submit", h.Discovery.SubmitEvent)

	inquiries := api.Group("/inquiries")
	inquiries.Use(limiter.New(limiter.Config{Max: 15, Expiration: time.Minute}))
	inquiries.Post("/", h.Inquiry.Submit)

	referrals := api.Group("/referrals")
	referrals.Use(limiter.New(limiter.Config{Max: 15, Expiration: time.Minute}))
	referrals.Post("/booking", h.Commerce.SubmitBookingReferral)

	api.Get("/platform/settings", h.Platform.Settings)

	reservations := api.Group("/reservations", middleware.OptionalAuth(h.JWT))
	reservations.Post("/hotel", h.Reservation.CreateHotel)
	reservations.Post("/table", h.Reservation.CreateTable)
	reservations.Get("/:id", h.Reservation.GetByID)

	orders := api.Group("/orders", middleware.OptionalAuth(h.JWT))
	orders.Post("/product", h.Order.CreateProduct)
	orders.Post("/food", h.Order.CreateFood)
	orders.Post("/discovery", h.Order.CreateDiscovery)
	orders.Get("/:id/track", h.Order.Track)

	rewards := api.Group("/rewards", middleware.Auth(h.JWT))
	rewards.Get("/balance", h.Commerce.UserRewardBalance)
	rewards.Post("/redeem", h.Commerce.UserRedeemRewards)

	api.Get("/alerts", h.Messaging.ListActiveAlerts)

	notifications := api.Group("/notifications", middleware.Auth(h.JWT))
	notifications.Get("/", h.Messaging.ListNotifications)
	notifications.Post("/read", h.Messaging.MarkNotificationsRead)

	chat := api.Group("/chat", middleware.Auth(h.JWT))
	chat.Post("/conversations", h.Messaging.StartChat)
	chat.Get("/conversations", h.Messaging.ListMyChats)
	chat.Get("/conversations/:id", h.Messaging.GetMyChat)
	chat.Post("/conversations/:id/messages", h.Messaging.SendMyMessage)

	api.Post("/webhooks/inbound", h.Messaging.InboundWebhook)
	api.Post("/webhooks/selcom", h.Payment.SelcomWebhook)

	payments := api.Group("/payments")
	payments.Get("/status", h.Payment.PaymentStatus)
	if h.Selcom.MockMode {
		payments.Post("/mock/complete", h.Payment.MockComplete)
		payments.Post("/mock/refund", h.Payment.MockRefund)
		payments.Post("/mock/expire-hold", h.Payment.MockExpireHold)
		payments.Post("/mock/reserve-capacity", h.Payment.MockReserveCapacity)
		payments.Post("/mock/expire-reservations", h.Payment.MockExpireReservations)
	}

	admin := api.Group("/admin", middleware.Auth(h.JWT), middleware.RequireAdmin())
	admin.Get("/me", h.Admin.Me)
	admin.Get("/dashboard", h.Admin.Dashboard)
	admin.Get("/analytics", h.Admin.Analytics)
	admin.Get("/store/:hotelId/dashboard", h.Admin.StoreDashboard)
	admin.Get("/hotels", h.Admin.ListHotels)
	admin.Post("/hotels", h.Admin.CreateHotel)
	admin.Get("/hotels/:id", h.Admin.GetHotel)
	admin.Patch("/hotels/:id", h.Admin.UpdateHotel)
	admin.Get("/hotels/:hotelId/products", h.Admin.ListProducts)
	admin.Post("/hotels/:hotelId/products", h.Admin.CreateProduct)
	admin.Get("/hotels/:hotelId/menu-items", h.Admin.ListMenuItems)
	admin.Post("/hotels/:hotelId/menu-items", h.Admin.CreateMenuItem)
	admin.Patch("/menu-items/:id", h.Admin.UpdateMenuItem)
	admin.Delete("/menu-items/:id", h.Admin.DeleteMenuItem)
	admin.Post("/hotels/:hotelId/import/:kind", h.Admin.ImportCSV)
	admin.Patch("/products/:id", h.Admin.UpdateProduct)
	admin.Get("/orders/summary", h.Admin.OrderSummary)
	admin.Get("/orders", h.Admin.ListOrders)
	admin.Get("/orders/:id", h.Admin.GetOrder)
	admin.Get("/guest-stays", h.Admin.ListGuestStays)
	admin.Patch("/orders/:id/status", h.Admin.UpdateOrderStatus)
	admin.Get("/reservations", h.Admin.ListReservations)
	admin.Get("/reservations/:id", h.Admin.GetReservation)
	admin.Patch("/reservations/:id/status", h.Admin.UpdateReservationStatus)
	admin.Get("/discovery", h.Discovery.AdminList)
	admin.Post("/discovery", h.Discovery.AdminCreate)
	admin.Get("/discovery/:id", h.Discovery.AdminGet)
	admin.Patch("/discovery/:id", h.Discovery.AdminUpdate)
	admin.Get("/alerts", h.Messaging.AdminListAlerts)
	admin.Post("/alerts", h.Messaging.AdminCreateAlert)
	admin.Patch("/alerts/:id", h.Messaging.AdminUpdateAlert)
	admin.Get("/chat", h.Messaging.AdminListChats)
	admin.Get("/chat/:id", h.Messaging.AdminGetChat)
	admin.Post("/chat/:id/messages", h.Messaging.AdminSendChatMessage)
	admin.Patch("/chat/:id/close", h.Messaging.AdminCloseChat)
	admin.Get("/webhooks", h.Messaging.AdminListWebhooks)
	admin.Post("/webhooks", h.Messaging.AdminCreateWebhook)
	admin.Patch("/webhooks/:id", h.Messaging.AdminUpdateWebhook)
	admin.Get("/webhooks/deliveries", h.Messaging.AdminListWebhookDeliveries)
	admin.Get("/inquiries", h.Inquiry.AdminList)
	admin.Get("/guest-requests", h.Platform.AdminListGuestRequests)
	admin.Patch("/inquiries/:id/status", h.Inquiry.AdminUpdateStatus)

	// Commerce engine (brief §3, §8) — admin only; commission rates never exposed to guests/partners.
	admin.Get("/influencers", h.Commerce.AdminListInfluencers)
	admin.Post("/influencers", h.Commerce.AdminCreateInfluencer)
	admin.Patch("/influencers/:id", h.Commerce.AdminUpdateInfluencer)
	admin.Get("/booking-referrals", h.Commerce.AdminListBookingReferrals)
	admin.Get("/suppliers", h.Commerce.AdminListSuppliers)
	admin.Post("/suppliers", h.Commerce.AdminCreateSupplier)
	admin.Patch("/suppliers/:id", h.Commerce.AdminUpdateSupplier)
	admin.Get("/catalogue", h.Commerce.AdminListCatalogue)
	admin.Post("/catalogue", h.Commerce.AdminCreateCatalogueItem)
	admin.Patch("/catalogue/:id", h.Commerce.AdminUpdateCatalogueItem)
	admin.Post("/catalogue/:id/visibility", h.Commerce.AdminSetCatalogueVisibility)
	admin.Get("/commission-rules", h.Commerce.AdminListCommissionRules)
	admin.Put("/commission-rules", h.Commerce.AdminUpsertCommissionRule)
	admin.Get("/commission-records", h.Commerce.AdminListCommissionRecords)
	admin.Get("/payout-batches", h.Commerce.AdminListPayoutBatches)
	admin.Post("/payout-batches/generate", h.Commerce.AdminGeneratePayoutBatches)
	admin.Post("/payout-batches/:id/release", h.Commerce.AdminReleasePayoutBatch)
	admin.Put("/reward-rules", h.Commerce.AdminUpsertRewardRule)
	admin.Get("/event-log", h.Commerce.AdminListEventLog)

	admin.Get("/platform/settings", h.Platform.AdminSettings)
	admin.Put("/platform/settings", h.Platform.AdminUpdateSettings)

	partner := api.Group("/partner", middleware.Auth(h.JWT), middleware.RequirePartner())
	partner.Get("/dashboard", h.Partner.Dashboard)
	partner.Get("/orders", h.Partner.ListOrders)
	partner.Get("/products", h.Partner.ListProducts)
	partner.Patch("/products/:id", h.Partner.UpdateProduct)
	partner.Get("/menu-items", h.Partner.ListMenuItems)
	partner.Post("/menu-items", h.Partner.CreateMenuItem)
	partner.Patch("/menu-items/:id", h.Partner.UpdateMenuItem)
	partner.Get("/guest-stays", h.Partner.ListGuestStays)
	partner.Get("/commissions", h.Partner.ListCommissions)
	partner.Get("/settings", h.Partner.GetSettings)
	partner.Patch("/settings", h.Partner.UpdateSettings)
	partner.Get("/referrals", h.Partner.ListReferrals)
	partner.Post("/referrals", h.Partner.CreateReferral)
}
