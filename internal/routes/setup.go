package routes

import (
	"github.com/AkashKesav/API2SDK/configs" // Added
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	app *fiber.App,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	collectionController *controllers.CollectionController,
	sdkController *controllers.SDKController,
	adminController *controllers.AdminController,
	healthController *controllers.HealthController,
	htmxController *controllers.HTMXController,
	publicApiController *controllers.PublicAPIController,
	authService services.AuthService,
	logger *zap.Logger,
	config *configs.Config, // Changed from jwtSecret string to *configs.Config
) {
	// API version 1 routes
	api := app.Group("/api/v1")

	// Auth routes
	authGroup := api.Group("/auth")

	// setupAuthRoutes will handle all auth routes including /profile
	setupAuthRoutes(authGroup, authController, config, logger)

	// Health check routes (public)
	setupHealthRoutes(api, healthController)

	// User self-service routes - protected
	usersGroup := api.Group("/users", middleware.JWTMiddleware(config, logger)) // Renamed and applied middleware here
	setupUserRoutes(usersGroup, userController)

	// Collection routes - protected
	collectionsGroup := api.Group("/collections", middleware.JWTMiddleware(config, logger)) // Renamed and applied middleware here
	setupCollectionRoutes(collectionsGroup, collectionController)

	// SDK generation routes - protected
	generateGroup := api.Group("/generate", middleware.JWTMiddleware(config, logger)) // Renamed and applied middleware here
	setupGeneratorRoutes(generateGroup, sdkController)

	// SDK management routes (history, deletion, download) - protected
	sdksGroup := api.Group("/sdks", middleware.JWTMiddleware(config, logger)) // Renamed and applied middleware here
	setupSDKRoutes(sdksGroup, sdkController)

	// Public API browsing routes (public - no auth required)
	publicApisGroup := api.Group("/public-apis") // Renamed for consistency
	setupPublicAPIRoutes(publicApisGroup, publicApiController)

	// HTMX routes - mixed public and protected
	// Create both public and protected HTMX groups
	publicHtmxGroup := api.Group("/htmx")
	protectedHtmxGroup := api.Group("/htmx", middleware.JWTMiddleware(config, logger))

	// Setup routes with appropriate protection levels
	setupHTMXRoutes(publicHtmxGroup, protectedHtmxGroup, htmxController)

	// Admin routes - require JWT + admin role
	// This will be api.Group("/admin", middleware.JWTMiddleware(config), middleware.AdminRequired())
	adminGroup := api.Group("/admin", middleware.JWTMiddleware(config, logger), middleware.AdminRequired()) // Renamed and applied middleware here
	setupAdminRoutes(adminGroup, adminController, userController)

	// Serve static files for frontend
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendFile("./web/index.html")
	})

	app.Get("/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		if path == "" {
			path = "index.html"
		}
		return c.SendFile("./web/" + path)
	})

	logger.Info("All routes configured successfully")
}
