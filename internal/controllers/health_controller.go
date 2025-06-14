package controllers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// HealthController handles health check requests
type HealthController struct {
	logger *zap.Logger
}

// NewHealthController creates a new HealthController
func NewHealthController(logger *zap.Logger) *HealthController {
	return &HealthController{
		logger: logger,
	}
}

// CheckHealth handles the health check endpoint
func (hc *HealthController) CheckHealth(c fiber.Ctx) error {
	hc.logger.Info("Health check requested")

	// Basic health check: service is running
	response := fiber.Map{
		"status":  "OK",
		"message": "API2SDK Service is running",
		"healthy": true,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
