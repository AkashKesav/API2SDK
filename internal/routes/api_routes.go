package routes

import (
	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(api fiber.Router, healthController *controllers.HealthController) {
	health := api.Group("/health")
	health.Get("/", healthController.CheckHealth)
}

// setupCollectionRoutes configures collection management endpoints
func setupCollectionRoutes(api fiber.Router, collectionController *controllers.CollectionController) {
	api.Post("/", collectionController.CreateCollection)
	api.Get("/", collectionController.GetUserCollections)
	api.Get("/:id", collectionController.GetCollectionByID)
	api.Put("/:id", collectionController.UpdateCollection)
	api.Delete("/:id", collectionController.DeleteCollection)
	api.Post("/:id/generate-openapi-spec", collectionController.GenerateOpenAPISpec)
}

// setupGeneratorRoutes configures SDK generation endpoints
func setupGeneratorRoutes(api fiber.Router, sdkController *controllers.SDKController) {
	api.Post("/sdk", sdkController.GenerateSDK)
	api.Post("/mcp", sdkController.GenerateMCP)
	api.Get("/languages", sdkController.GetSupportedLanguages)
}

// setupSDKRoutes configures SDK management endpoints (history, deletion, download)
func setupSDKRoutes(api fiber.Router, sdkController *controllers.SDKController) {
	api.Get("/", sdkController.GetSDKHistory)
	api.Delete("/:id", sdkController.DeleteSDK)
	api.Get("/:id/download", sdkController.DownloadSDK)
}

// setupPublicAPIRoutes configures public API browsing routes
func setupPublicAPIRoutes(api fiber.Router, publicAPIController *controllers.PublicAPIController) {
	api.Get("/", publicAPIController.GetPublicAPIs)
	api.Get("/popular", publicAPIController.GetPopularAPIs)
	api.Get("/categories", publicAPIController.GetCategories)
	api.Get("/:id", publicAPIController.GetPublicAPIByID)
}

// setupHTMXRoutes configures HTMX specific routes
func setupHTMXRoutes(publicApi fiber.Router, protectedApi fiber.Router, htmxController *controllers.HTMXController) {
	// Public HTMX routes
	publicApi.Get("/framework-options", controllers.GetFrameworkOptionsHTML)
	publicApi.Get("/popular-apis", htmxController.GetPopularAPIsHTML)
	publicApi.Get("/theme-toggle", controllers.GetThemeToggleHTML)
	publicApi.Post("/theme-toggle", controllers.HandleThemeToggle)

	// Protected HTMX routes (now with no-auth middleware)
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
func setupAuthRoutes(api fiber.Router, authController *controllers.AuthController, config *configs.Config, logger *zap.Logger) {
	logger.Info("Setting up auth routes", zap.String("group", "auth"))

	if authController == nil {
		logger.Error("AuthController is nil!")
		return
	}

	api.Post("/register", authController.Register)
	api.Post("/login", authController.Login)
	api.Get("/profile", authController.GetUserProfile)
	api.Post("/logout", authController.Logout)

	logger.Info("Auth routes setup completed", zap.Int("routes_added", 4))
}

// setupUserRoutes configures user management endpoints (self-service)
func setupUserRoutes(api fiber.Router, userController *controllers.UserController) {
	api.Get("/me", userController.GetMe)
	api.Put("/me", userController.UpdateMe)
}

// setupUserMCPRoutes configures user MCP management endpoints
func setupUserMCPRoutes(api fiber.Router, userMCPController *controllers.UserMCPController) {
	mcps := api.Group("/mcps")
	mcps.Post("/", userMCPController.CreateMCPInstance)
	mcps.Get("/", userMCPController.ListMCPInstances)
	mcps.Delete("/:instanceID", userMCPController.DeleteMCPInstance)
	mcps.Get("/:instanceID/resources", userMCPController.ListResources)
}

// setupAdminRoutes configures admin-specific endpoints
func setupAdminRoutes(api fiber.Router, adminController *controllers.AdminController, userController *controllers.UserController) {
	// User management by admin
	api.Get("/users", userController.GetAllUsers)
	api.Get("/users/:id", userController.GetUserByID)
	api.Put("/users/:id", userController.UpdateUser)
	api.Delete("/users/:id", userController.DeleteUser)

	// Platform settings by admin
	api.Get("/settings", adminController.GetPlatformSettings)
	api.Put("/settings", adminController.UpdatePlatformSettings)
}
