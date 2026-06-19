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
	Payment        *handler.PaymentHandler
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
	hotels.Get("/slug/:slug/rooms", h.Hotel.ListRooms)
	hotels.Get("/slug/:slug/menu", h.Hotel.ListMenu)
	hotels.Post("/slug/:slug/scan", middleware.OptionalAuth(h.JWT), h.Hotel.RecordScan)
	hotels.Get("/slug/:slug/discovery", h.Discovery.PortalBySlug)
	hotels.Get("/:code/products", h.Hotel.ListProductsByCode)
	hotels.Get("/:code", h.Hotel.GetByCode)

	discovery := api.Group("/discovery")
	discovery.Post("/events/submit", h.Discovery.SubmitEvent)

	reservations := api.Group("/reservations", middleware.OptionalAuth(h.JWT))
	reservations.Post("/hotel", h.Reservation.CreateHotel)
	reservations.Post("/table", h.Reservation.CreateTable)
	reservations.Get("/:id", h.Reservation.GetByID)

	orders := api.Group("/orders", middleware.OptionalAuth(h.JWT))
	orders.Post("/product", h.Order.CreateProduct)
	orders.Post("/food", h.Order.CreateFood)
	orders.Get("/:id/track", h.Order.Track)

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
}
