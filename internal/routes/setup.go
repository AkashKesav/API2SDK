package routes

import (
	"time"

	"github.com/AkashKesav/API2SDK/configs"
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
	mcpController *controllers.MCPController,
	userMCPController *controllers.UserMCPController,
	authService services.AuthService,
	logger *zap.Logger,
	config *configs.Config,
) {
	// API version 1 routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	authGroup := api.Group("/auth")
	setupAuthRoutes(authGroup, authController, config, logger)

	// Health check routes (public)
	setupHealthRoutes(api, healthController)

	// User self-service routes
	usersGroup := api.Group("/users", middleware.NoAuthMiddleware())
	setupUserRoutes(usersGroup, userController)
	setupUserMCPRoutes(usersGroup, userMCPController)

	// Collection routes
	collectionsGroup := api.Group("/collections", middleware.NoAuthMiddleware())
	setupCollectionRoutes(collectionsGroup, collectionController)

	// SDK generation routes with rate limiting
	generateGroup := api.Group("/generate",
		middleware.NoAuthMiddleware(),
		middleware.EnhancedRateLimitMiddleware(middleware.NewRateLimiter(10, time.Minute), logger),
		middleware.CircuitBreakerMiddleware("sdk_generation", logger))
	setupGeneratorRoutes(generateGroup, sdkController)

	// SDK management routes (history, deletion, download)
	sdksGroup := api.Group("/sdks", middleware.NoAuthMiddleware())
	setupSDKRoutes(sdksGroup, sdkController)

	// Public API browsing routes (public - no auth required)
	publicApisGroup := api.Group("/public-apis")
	setupPublicAPIRoutes(publicApisGroup, publicApiController)

	// HTMX routes - mixed public and protected
	publicHtmxGroup := api.Group("/htmx")
	protectedHtmxGroup := api.Group("/htmx", middleware.NoAuthMiddleware())
	setupHTMXRoutes(publicHtmxGroup, protectedHtmxGroup, htmxController)

	// Admin routes
	adminGroup := api.Group("/admin", middleware.NoAuthMiddleware(), middleware.AdminRequired())
	setupAdminRoutes(adminGroup, adminController, userController)

	// MCP routes
	mcpGroup := app.Group("/mcp")
	mcpRouter := NewMCPRouter(mcpGroup, mcpController)
	mcpRouter.SetupRoutes()

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
