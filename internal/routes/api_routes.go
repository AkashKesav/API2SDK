package routes

import (
	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/AkashKesav/API2SDK/internal/middleware" // Re-add middleware import
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// setupHealthRoutes configures health check endpoints
// It now takes a pointer to HealthController as passed from setup.go
func setupHealthRoutes(api fiber.Router, healthController *controllers.HealthController) {
	health := api.Group("/health")
	health.Get("/", healthController.CheckHealth)
}

// setupCollectionRoutes configures collection management endpoints
// JWTMiddleware is applied in SetupAPIRoutes (in setup.go) before calling this.
func setupCollectionRoutes(api fiber.Router, collectionController *controllers.CollectionController) {
	// collections := api.Group("/collections") // Group is already created and passed as `api` argument
	api.Post("/", collectionController.CreateCollection)
	api.Get("/", collectionController.GetUserCollections)
	api.Get("/:id", collectionController.GetCollectionByID)
	api.Put("/:id", collectionController.UpdateCollection)
	api.Delete("/:id", collectionController.DeleteCollection)
	api.Post("/:id/generate-openapi-spec", collectionController.GenerateOpenAPISpec) // Renamed route and method
	// api.Post("/from-public-api", collectionController.CreateCollectionFromPublicAPI) // Method commented out in controller
}

// setupGeneratorRoutes configures SDK generation endpoints
// JWTMiddleware is applied in SetupAPIRoutes (in setup.go) before calling this.
func setupGeneratorRoutes(api fiber.Router, sdkController *controllers.SDKController) {
	// generator := api.Group("/generate") // Group is already created and passed as `api` argument
	api.Post("/sdk", sdkController.GenerateSDK)
	api.Get("/languages", sdkController.GetSupportedLanguages)
}

// setupSDKRoutes configures SDK management endpoints (history, deletion, download)
// JWTMiddleware is applied in SetupAPIRoutes (in setup.go) before calling this.
func setupSDKRoutes(api fiber.Router, sdkController *controllers.SDKController) {
	// sdks := api.Group("/sdks") // Group is already created and passed as `api` argument
	api.Get("/", sdkController.GetSDKHistory)           // GET /api/v1/sdks - get user's SDK history with pagination
	api.Delete("/:id", sdkController.DeleteSDK)         // DELETE /api/v1/sdks/:id - soft delete an SDK
	api.Get("/:id/download", sdkController.DownloadSDK) // GET /api/v1/sdks/:id/download - download SDK files
}

// Public API browsing routes
// JWTMiddleware is applied in SetupAPIRoutes (in setup.go) if needed, or handled per route.
func setupPublicAPIRoutes(api fiber.Router, publicAPIController *controllers.PublicAPIController) {
	// publicAPI := api.Group("/public-apis") // Group is already created and passed as `api` argument
	api.Get("/", publicAPIController.GetPublicAPIs)
	api.Get("/popular", publicAPIController.GetPopularAPIs)
	api.Get("/categories", publicAPIController.GetCategories)
	api.Get("/:id", publicAPIController.GetPublicAPIByID)
	// Example for a protected route within this group (middleware would be applied in SetupAPIRoutes or specifically here)
	// api.Post("/", middleware.JWTMiddleware(yourSecret), publicAPIController.CreatePublicAPI)
}

// HTMX specific routes
// Some routes are public, some require authentication
func setupHTMXRoutes(publicApi fiber.Router, protectedApi fiber.Router, htmxController *controllers.HTMXController) {
	// Public HTMX routes (no authentication required)
	publicApi.Get("/framework-options", controllers.GetFrameworkOptionsHTML)
	publicApi.Get("/popular-apis", htmxController.GetPopularAPIsHTML)
	publicApi.Get("/theme-toggle", controllers.GetThemeToggleHTML)
	publicApi.Post("/theme-toggle", controllers.HandleThemeToggle)

	// Protected HTMX routes (authentication required)
	protectedApi.Post("/collections", htmxController.CreateCollectionHTML)
	protectedApi.Post("/collections/from-url", htmxController.CreateCollectionFromURLHTML)
	protectedApi.Post("/collections/from-public-api", htmxController.CreateCollectionFromPublicAPIHTML)
	protectedApi.Get("/sdk-history", htmxController.GetSDKHistoryHTML)
	protectedApi.Delete("/sdks/:id", htmxController.DeleteSDKHTML)
	protectedApi.Get("/generation-status/:taskID", controllers.GetGenerationStatusHTML)
	protectedApi.Post("/cancel-generation/:taskID", controllers.CancelGenerationTaskHTML)
	protectedApi.Get("/user-profile-card", htmxController.GetUserProfileCardHTML)
}

// setupAuthRoutes configures authentication endpoints
func setupAuthRoutes(api fiber.Router, authController *controllers.AuthController, config *configs.Config, logger *zap.Logger) { // Added logger
	logger.Info("Setting up auth routes", zap.String("group", "auth"))

	// Test if the auth controller is valid
	if authController == nil {
		logger.Error("AuthController is nil!")
		return
	}

	api.Post("/register", authController.Register)                                                       // Public
	api.Post("/login", authController.Login)                                                             // Public
	api.Post("/refresh", middleware.RefreshTokenMiddleware(config, logger), authController.RefreshToken) // Requires refresh token

	// Define /profile route with JWT middleware - using Use() method (Fiber v3 compatible)
	jwtMiddleware := middleware.JWTMiddleware(config, logger)
	api.Use("/profile", jwtMiddleware)
	api.Get("/profile", authController.GetUserProfile)

	// Define /logout route with JWT middleware
	api.Use("/logout", jwtMiddleware)
	api.Post("/logout", authController.Logout)

	logger.Info("Auth routes setup completed", zap.Int("routes_added", 5))
}

// setupUserRoutes configures user management endpoints (self-service)
// JWTMiddleware is applied in SetupAPIRoutes (in setup.go) before calling this.
func setupUserRoutes(api fiber.Router, userController *controllers.UserController) {
	// user := api.Group("/users") // Group is already created and passed as `api` argument
	api.Get("/me", userController.GetMe)
	api.Put("/me", userController.UpdateMe)
	// Add other self-service user routes here if needed
	// e.g., api.Patch("/me/password", userController.ChangePassword)
}

// setupAdminRoutes configures admin-specific endpoints
// JWTMiddleware and AdminRequired middleware are applied in SetupAPIRoutes (in setup.go) before calling this.
func setupAdminRoutes(api fiber.Router, adminController *controllers.AdminController, userController *controllers.UserController) {
	// admin := api.Group("/admin") // Group is already created and passed as `api` argument

	// User management by admin (routes are on the `admin` group passed as `api`)
	api.Get("/users", userController.GetAllUsers)
	api.Get("/users/:id", userController.GetUserByID)
	api.Put("/users/:id", userController.UpdateUser)
	api.Delete("/users/:id", userController.DeleteUser)

	// Platform settings by admin
	api.Get("/settings", adminController.GetPlatformSettings)
	api.Put("/settings", adminController.UpdatePlatformSettings)

	// Other admin functions can be added here
	// e.g., api.Get("/stats", adminController.GetStats)
}
