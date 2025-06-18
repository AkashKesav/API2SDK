package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/routes"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	// Initialize Logger
	zapLogger, err := zap.NewDevelopment() // Or zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer zapLogger.Sync() // flushes buffer, if any

	appConfigs, err := configs.LoadConfig()
	if err != nil {
		zapLogger.Fatal("Failed to load config", zap.Error(err))
	}
	configs.InitConfig(appConfigs) // Initialize global config for accessors like GetPostmanAPIKey()

	// Initialize Database Connection
	if err := configs.InitDatabase(appConfigs); err != nil { // Assuming InitDatabase takes *configs.Config
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	db := configs.GetDatabase() // Get the database instance
	if db == nil {
		zapLogger.Fatal("Database instance is nil after initialization")
	}
	zapLogger.Info("Database initialized successfully")

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(db)                              // Removed zapLogger
	collectionRepo := repositories.NewCollectionRepository(db)                  // Renamed and removed zapLogger
	platformSettingsRepo := repositories.NewMongoPlatformSettingsRepository(db) // Renamed and removed zapLogger
	sdkRepo := repositories.NewSDKRepository(db, zapLogger)                     // Instantiate SDKRepository

	// Initialize Services
	authService := services.NewAuthService(userRepo, zapLogger, appConfigs)             // Pass the whole config object
	var platformSettingsService services.PlatformSettingsService                        // Interface type
	platformSettingsService = services.NewPlatformSettingsService(platformSettingsRepo) // Concrete type
	userService := services.NewUserService(userRepo, platformSettingsService, zapLogger)

	// Create PostmanClient for SDKService
	postmanClient := services.NewPostmanClient()

	// Get OpenAPI Generator path from environment or use default
	openAPIGenPath := os.Getenv("OPENAPI_GENERATOR_CLI_JAR")
	if openAPIGenPath == "" {
		openAPIGenPath = "openapi-generator-cli.jar" // Default
	}

	// Get mongo client from the database connection for SDKService
	mongoClient := db.Client()

	sdkService, err := services.NewSDKService(
		sdkRepo,
		mongoClient,
		appConfigs.MongoDBName,
		postmanClient,
		zapLogger,
		openAPIGenPath,
		services.GetPyGenScript(),
		services.GetPhpGenScript(),
		services.GetPhpVendorZip(),
	)
	if err != nil {
		zapLogger.Fatal("Failed to initialize SDK service", zap.Error(err))
	}

	collectionService := services.NewCollectionService(collectionRepo, zapLogger, sdkRepo, sdkService)

	// Use configs.GetPostmanAPIKey() to get the key from the initialized global config
	postmanAPIKey := configs.GetPostmanAPIKey()
	if postmanAPIKey == "" {
		zapLogger.Warn("POSTMAN_API_KEY is not set in environment. Postman API features will be limited.")
	}
	postmanAPIService := services.NewPostmanAPIService(zapLogger, postmanAPIKey) // Pass the key

	publicApiService := services.NewPublicAPIService(zapLogger, postmanAPIService, collectionService) // Pass PostmanAPIService and CollectionService

	// Initialize Controllers
	authController := controllers.NewAuthController(authService, userService, zapLogger)
	userController := controllers.NewUserController(userService, zapLogger)
	collectionController := controllers.NewCollectionController(collectionService, zapLogger)
	adminController := controllers.NewAdminController(userService, platformSettingsService, zapLogger)
	healthController := controllers.NewHealthController(zapLogger)
	htmxController := controllers.NewHTMXController(zapLogger, sdkService, sdkRepo, collectionService, postmanAPIService, publicApiService) // Added publicApiService
	sdkController := controllers.NewSDKController(sdkService, collectionService, zapLogger)
	// Pass postmanAPIService to PublicAPIController if it needs to interact with Postman API directly
	// For now, assuming PublicAPIController uses PublicAPIService which encapsulates Postman interactions.
	publicApiController := controllers.NewPublicAPIController(publicApiService, zapLogger)
	// postmanPublicApiController := controllers.NewPostmanPublicAPIController(postmanAPIService, zapLogger, collectionService, sdkService) // Added sdkService - Commented out as functionality might be covered by HTMXController + PublicAPIService

	// Create a new Fiber instance
	app := fiber.New(fiber.Config{
		ReadBufferSize:  32768,
		WriteBufferSize: 32768,
		BodyLimit:       100 * 1024 * 1024,
		ReadTimeout:     180 * time.Second,
		WriteTimeout:    180 * time.Second,
		IdleTimeout:     300 * time.Second,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := err.Error()

			if strings.Contains(message, "timeout") || strings.Contains(message, "Timeout") {
				code = fiber.StatusRequestTimeout
				message = "Request timeout - operation took too long to complete. Please try again or use a smaller collection."
			}

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			zapLogger.Error("Error handled", zap.Int("code", code), zap.String("message", message), zap.Error(err))

			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": message,
				"code":    code,
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\\\\n",
	}))

	// Determine allowed origins
	// In a real production app, you might get this from config or another env var
	allowedOrigins := []string{"http://localhost:3001"} // Default frontend dev port
	if appConfigs != nil && appConfigs.Port != "" {
		backendOrigin := fmt.Sprintf("http://localhost:%s", appConfigs.Port)
		allowedOrigins = append(allowedOrigins, backendOrigin)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins, // Specify origins
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodHead, fiber.MethodPut, fiber.MethodDelete, fiber.MethodPatch, fiber.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Serve static files from the "web" directory
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendFile("./web/index.html")
	})

	// Authentication pages
	app.Get("/login", func(c fiber.Ctx) error {
		return c.SendFile("./web/login.html")
	})

	app.Get("/register", func(c fiber.Ctx) error {
		return c.SendFile("./web/register.html")
	})

	// Serve assets
	app.Get("/assets/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		return c.SendFile("./web/assets/" + path)
	})

	// Health check endpoint
	app.Get("/health", healthController.CheckHealth) // Using the initialized healthController

	// Setup routes
	routes.SetupRoutes(app,
		authController,
		userController,
		collectionController,
		sdkController,
		adminController,
		healthController, // Pass healthController
		htmxController,
		publicApiController, // Pass publicApiController
		// postmanPublicApiController, // Commented out
		authService,
		zapLogger,
		appConfigs, // Pass the whole config object instead of just JWTSecret
	)

	// Start server
	port := appConfigs.Port // Use port from config
	if port == "" {
		port = "8080" // Fallback default port
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		zapLogger.Info("Gracefully shutting down...")
		if err := app.Shutdown(); err != nil {
			zapLogger.Error("Server shutdown failed", zap.Error(err))
		}
	}()

	zapLogger.Info("🚀 Server starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		zapLogger.Fatal("Server failed to start", zap.Error(err))
	}
}
