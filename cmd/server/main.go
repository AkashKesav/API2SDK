package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/AkashKesav/API2SDK/internal/mcp"
	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/routes"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// StdioRequest defines the structure for incoming stdio requests.
type StdioRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// ListToolsParams defines the parameters for the list_tools method.
type ListToolsParams struct {
	InstanceID string `json:"instanceId"`
}

// CallToolParams defines the parameters for the call_tool method.
type CallToolParams struct {
	InstanceID string                 `json:"instanceId"`
	Name       string                 `json:"name"`
	Arguments  map[string]interface{} `json:"arguments"`
}

// ListResourcesParams defines the parameters for the list_resources method.
type ListResourcesParams struct {
	InstanceID string `json:"instanceId"`
}

// StdioResponse defines the structure for stdio responses.
type StdioResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func main() {
	transport := flag.String("transport", "http", "transport layer to use: http or stdio")
	flag.Parse()

	// Initialize Logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer zapLogger.Sync()

	appConfigs, err := configs.LoadConfig()
	if err != nil {
		zapLogger.Fatal("Failed to load config", zap.Error(err))
	}
	configs.InitConfig(appConfigs)

	zapLogger.Info("Starting API2SDK server with MongoDB functionality")

	// Initialize Database Connection - required for operation
	if err := configs.InitDatabase(appConfigs); err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	zapLogger.Info("Database initialized successfully")

	// Initialize Repositories
	db := configs.GetDatabase()
	userRepo := repositories.NewUserRepository(db)
	collectionRepo := repositories.NewCollectionRepository(db)
	platformSettingsRepo := repositories.NewMongoPlatformSettingsRepository(db)
	sdkRepo := repositories.NewSDKRepository(db, zapLogger)
	integrationRepo := repositories.NewIntegrationRepository(db)
	mcpInstanceRepo := repositories.NewMCPInstanceRepository(db)
	zapLogger.Info("All repositories initialized with database")

	// Initialize Services
	authService := services.NewAuthService(userRepo, zapLogger, appConfigs)
	var platformSettingsService services.PlatformSettingsService = services.NewPlatformSettingsService(platformSettingsRepo)
	userService := services.NewUserService(userRepo, platformSettingsService, zapLogger)
	integrationService := services.NewIntegrationService(integrationRepo)
	toolProvider := services.NewToolProviderService(integrationService)
	mcpInstanceService := services.NewMCPInstanceService(mcpInstanceRepo, integrationService, toolProvider)
	mcpManager := mcp.NewMCPManager(zapLogger, integrationService, toolProvider)
	zapLogger.Info("All services initialized with database")

	// Create PostmanClient for SDKService
	postmanClient := services.NewPostmanClient(appConfigs)

	// Get OpenAPI Generator path from environment or use default
	openAPIGenPath := os.Getenv("OPENAPI_GENERATOR_CLI_JAR")
	if openAPIGenPath == "" {
		openAPIGenPath = "openapi-generator-cli.jar"
	}

	// Initialize SDK service
	sdkService, err := services.NewSDKService(
		sdkRepo,
		postmanClient,
		zapLogger,
		openAPIGenPath,
		services.PyGenScript,
		services.PhpGenScript,
		services.PhpVendorZip,
	)
	if err != nil {
		zapLogger.Fatal("Failed to initialize SDK service", zap.Error(err))
	}

	// Initialize collection service
	collectionService := services.NewCollectionService(collectionRepo, zapLogger, sdkService)

	// Use configs.GetPostmanAPIKey() to get the key from the initialized global config
	postmanAPIKey := configs.GetPostmanAPIKey()
	if postmanAPIKey == "" {
		zapLogger.Warn("POSTMAN_API_KEY is not set in environment. Postman API features will be limited.")
	}
	postmanAPIService := services.NewPostmanAPIService(zapLogger, postmanAPIKey)

	// Initialize public API service
	publicApiService := services.NewPublicAPIService(zapLogger, postmanAPIService, collectionService, db)

	// Initialize Controllers
	healthController := controllers.NewHealthController(zapLogger)
	authController := controllers.NewAuthController(authService, userService, zapLogger)
	userController := controllers.NewUserController(userService, zapLogger)
	collectionController := controllers.NewCollectionController(collectionService, zapLogger)
	adminController := controllers.NewAdminController(userService, services.NewPlatformSettingsService(platformSettingsRepo), sdkService, collectionService, zapLogger)
	sdkController := controllers.NewSDKController(sdkService, collectionService, services.NewPlatformSettingsService(platformSettingsRepo), zapLogger)
	htmxController := controllers.NewHTMXController(zapLogger, collectionService, postmanAPIService, publicApiService)
	publicApiController := controllers.NewPublicAPIController(publicApiService, zapLogger)
	mcpController := controllers.NewMCPController(mcpInstanceService, integrationService, mcpManager, zapLogger)
	userMCPController := controllers.NewUserMCPController(mcpInstanceService, integrationService)

	if *transport == "stdio" {
		zapLogger.Info("Starting server in stdio mode")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Bytes()
			var req StdioRequest
			if err := json.Unmarshal(line, &req); err != nil {
				response := StdioResponse{Error: "Invalid JSON request"}
				respBytes, _ := json.Marshal(response)
				fmt.Println(string(respBytes))
				continue
			}

			var resp StdioResponse
			switch req.Method {
			case "list_tools":
				var params ListToolsParams
				if err := json.Unmarshal(req.Params, &params); err != nil {
					resp.Error = "Invalid params for list_tools"
				} else {
					objId, err := primitive.ObjectIDFromHex(params.InstanceID)
					if err != nil {
						resp.Error = "Invalid instance ID"
					} else {
						tools, err := mcpInstanceService.GetTools(context.Background(), objId)
						if err != nil {
							resp.Error = err.Error()
						} else {
							resp.Result = tools
						}
					}
				}
			case "call_tool":
				var params CallToolParams
				if err := json.Unmarshal(req.Params, &params); err != nil {
					resp.Error = "Invalid params for call_tool"
				} else {
					objId, err := primitive.ObjectIDFromHex(params.InstanceID)
					if err != nil {
						resp.Error = "Invalid instance ID"
					} else {
						result, err := mcpInstanceService.ExecuteToolCall(context.Background(), objId, params.Name, params.Arguments)
						if err != nil {
							resp.Error = err.Error()
						} else {
							resp.Result = result
						}
					}
				}
			case "list_resources":
				var params ListResourcesParams
				if err := json.Unmarshal(req.Params, &params); err != nil {
					resp.Error = "Invalid params for list_resources"
				} else {
					objId, err := primitive.ObjectIDFromHex(params.InstanceID)
					if err != nil {
						resp.Error = "Invalid instance ID"
					} else {
						resources, err := mcpInstanceService.GetResources(context.Background(), objId)
						if err != nil {
							resp.Error = err.Error()
						} else {
							resp.Result = resources
						}
					}
				}
			default:
				resp.Error = "Unknown method: " + req.Method
			}

			respBytes, _ := json.Marshal(resp)
			fmt.Println(string(respBytes))
		}

		if err := scanner.Err(); err != nil {
			zapLogger.Error("Error reading from stdin", zap.Error(err))
		}
	} else {
		// Create Fiber app
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

		// Enhanced Middleware
		app.Use(recover.New())
		app.Use(middleware.AuthSkipMiddleware())
		app.Use(middleware.DefaultSecurityHeadersMiddleware())
		// Disable input validation in development
		if !appConfigs.IsDevelopment() {
			app.Use(middleware.DefaultInputValidationMiddleware())
		}
		app.Use(middleware.FileValidationMiddleware())
		app.Use(middleware.CSRFTokenGeneratorMiddleware())
		app.Use(middleware.TracingMiddleware(zapLogger))
		app.Use(middleware.MetricsMiddleware(zapLogger))
		app.Use(middleware.ErrorHandlerMiddleware(middleware.ErrorHandlerConfig{
			Logger:           zapLogger,
			EnableStackTrace: appConfigs.Environment == "development",
			EnableDebugMode:  appConfigs.Environment == "development",
		}))
		app.Use(logger.New(logger.Config{
			Format: "[${ip}]:${port} ${status} - ${method} ${path} - ${locals:trace_context}\\n",
		}))

		// CORS
		app.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080"},
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

		// Setup all application routes
		routes.SetupRoutes(
			app,
			authController,
			userController,
			collectionController,
			sdkController,
			adminController,
			healthController,
			htmxController,
			publicApiController,
			mcpController,
			userMCPController,
			authService,
			zapLogger,
			appConfigs,
		)

		// Graceful Shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-quit
			zapLogger.Info("Shutting down server...")
			if err := app.Shutdown(); err != nil {
				zapLogger.Fatal("Server forced to shutdown:", zap.Error(err))
			}
		}()

		port := appConfigs.Port
		if port == "" {
			port = "8080"
		}
		zapLogger.Info("🚀 Server starting", zap.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}

		zapLogger.Info("Server exiting")
	}
}
