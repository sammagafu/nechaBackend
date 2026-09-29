package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/pkg/response"
	"gorm.io/gorm"
)

func Health(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{"status": "ok"})
}

func HealthReady(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unavailable",
				"db":     "error",
			})
		}
		if err := sqlDB.PingContext(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unavailable",
				"db":     "down",
			})
		}
		return response.OK(c, fiber.Map{"status": "ready", "db": "ok"})
	}
}
