package controllers

import (
	"fmt"
	"html"
	"html/template"
	"path/filepath"
	"sync"
	"time" // For cookie expiration in HandleThemeToggle

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type HTMXController struct {
	logger            *zap.Logger
	collectionService *services.CollectionService
	postmanAPIService *services.PostmanAPIService
	publicAPIService  *services.PublicAPIService // Added PublicAPIService
	templates         *template.Template
}

var (
	generationStatusTemplate    *template.Template
	onceStatus                  sync.Once
	cancelGenerationTemplate    *template.Template
	onceCancel                  sync.Once
	themeToggleTemplate         *template.Template
	onceTheme                   sync.Once
	themeToggleResponseTemplate *template.Template
	onceThemeResponse           sync.Once
)

func NewHTMXController(logger *zap.Logger, collectionService *services.CollectionService, postmanAPIService *services.PostmanAPIService, publicAPIService *services.PublicAPIService) *HTMXController {
	templates, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		logger.Fatal("Failed to parse HTML templates", zap.Error(err))
	}

	return &HTMXController{
		logger:            logger,
		collectionService: collectionService,
		postmanAPIService: postmanAPIService,
		publicAPIService:  publicAPIService, // Store injected PublicAPIService
		templates:         templates,
	}
}

// GetSDKHistoryHTML returns HTML fragment for SDK history
func (hc *HTMXController) GetSDKHistoryHTML(c fiber.Ctx) error {
	hc.logger.Info("GetSDKHistoryHTML called")
	c.Set("Content-Type", "text/html")
	err := hc.templates.ExecuteTemplate(c.Response().BodyWriter(), "sdk_history.html", nil)
	if err != nil {
		hc.logger.Error("Failed to execute template sdk_history.html", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// GetFrameworkOptionsHTML returns framework options based on language
// This can remain a standalone function as it doesn't depend on controller state.
func GetFrameworkOptionsHTML(c fiber.Ctx) error {
	language := c.FormValue("language")
	if language == "" {
		language = c.Query("language")
	}

	frameworks := map[string][]string{
		"javascript": {"axios", "fetch", "node-fetch"},
		"typescript": {"axios", "fetch", "node-fetch"},
		"python":     {"requests", "httpx", "urllib"},
		"go":         {"net/http", "resty", "gorequest"},
		"java":       {"httpclient", "okhttp", "apache-httpclient"},
		"swift":      {"urlsession", "alamofire"},
		"kotlin":     {"okhttp", "ktor", "retrofit"},
		"csharp":     {"httpclient", "restsharp"},
		"c#":         {"httpclient", "restsharp"},
		"php":        {"guzzle", "curl", "file_get_contents"},
		"ruby":       {"net-http", "httparty", "faraday"},
	}

	options, exists := frameworks[language]
	if !exists {
		return c.SendString("")
	}

	html := ""
	for _, framework := range options {
		html += fmt.Sprintf(`<option value="%s">%s</option>`, framework, framework)
	}

	return c.SendString(html)
}

// DeleteSDKHTML handles SDK deletion and returns empty content
func (hc *HTMXController) DeleteSDKHTML(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).SendString(`
<div class="alert alert-error">
<i class="fas fa-exclamation-circle"></i>
SDK ID is required
</div>
`)
	}
	hc.logger.Info("SDK Deletion requested via HTMX (currently placeholder)", zap.String("sdkID", id))

	// Sanitize the ID to prevent XSS
	escapedID := html.EscapeString(id)

	c.Set("Content-Type", "text/html")
	err := hc.templates.ExecuteTemplate(c.Response().BodyWriter(), "sdk_delete.html", fiber.Map{"ID": escapedID})
	if err != nil {
		hc.logger.Error("Failed to execute template sdk_delete.html", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// GetPopularAPIsHTML returns HTML fragment for popular APIs
func (hc *HTMXController) GetPopularAPIsHTML(c fiber.Ctx) error {
	// Use the injected publicAPIService
	apis := hc.publicAPIService.GetPopularAPIs()

	c.Set("Content-Type", "text/html")
	err := hc.templates.ExecuteTemplate(c.Response().BodyWriter(), "popular_apis.html", fiber.Map{"APIs": apis})
	if err != nil {
		hc.logger.Error("Failed to execute template popular_apis.html", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// CreateCollectionHTML handles collection creation from HTMX forms and returns JSON
func (hc *HTMXController) CreateCollectionHTML(c fiber.Ctx) error {
	var req models.CreateCollectionRequest

	// Handle form submission from HTMX
	req.Name = c.FormValue("name")
	req.Description = c.FormValue("description")

	// Check if file was uploaded
	file, err := c.FormFile("file")
	if err == nil && file != nil {
		// Read file content
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Failed to read uploaded file: " + err.Error(),
			})
		}
		defer fileContent.Close()

		// Read file as string
		fileBytes := make([]byte, file.Size)
		_, err = fileContent.Read(fileBytes)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Failed to read file content: " + err.Error(),
			})
		}

		req.PostmanData = string(fileBytes)
		if req.Name == "" {
			req.Name = file.Filename
		}
	} else {
		// Check for JSON data in form
		postmanData := c.FormValue("postman_data")
		if postmanData != "" {
			req.PostmanData = postmanData
		}
	}

	if req.Name == "" {
		req.Name = "Untitled Collection"
	}

	if req.PostmanData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Postman collection data is required (either upload a file or paste JSON).",
		})
	}

	// Get user ID from middleware context
	userIDStr := "60d5ec49e79c9e001a8d0b1a" // Hardcoded user ID

	// Use the injected collectionService
	collection, err := hc.collectionService.CreateCollection(&req, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create collection: " + err.Error(),
		})
	}

	// Generate OpenAPI spec if PostmanData was provided
	if req.PostmanData != "" {
		_, specContent, err := hc.collectionService.GenerateOpenAPISpec(collection.ID.Hex())
		if err != nil {
			hc.logger.Error("Failed to generate OpenAPI spec after collection creation", zap.Error(err), zap.String("collectionID", collection.ID.Hex()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Collection created (ID: %s), but failed to generate OpenAPI spec: %s", collection.ID.Hex(), err.Error()),
				"data": fiber.Map{
					"collection_id": collection.ID.Hex(),
				},
			})
		}

		// Return success JSON with data for the OpenAPI preview step
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Collection uploaded and OpenAPI spec generated successfully!",
			"data": fiber.Map{
				"collection_id":   collection.ID.Hex(),
				"openapi_spec":    specContent,
				"collection_name": collection.Name,
			},
		})
	} else {
		// Return success JSON for config step
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Collection created successfully!",
			"data": fiber.Map{
				"collection_id":   collection.ID.Hex(),
				"collection_name": collection.Name,
			},
		})
	}
}

// CreateCollectionFromURLHTML handles collection creation from Postman URL and returns JSON
func (hc *HTMXController) CreateCollectionFromURLHTML(c fiber.Ctx) error {
	postmanURL := c.FormValue("postman_url")
	collectionName := c.FormValue("collection_name") // User-provided name for the collection

	if postmanURL == "" {
		// This response should be handled by the htmx:afterRequest in JS to show an error in a modal or specific div
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Postman URL is required.",
		})
	}

	// Get user ID from middleware context
	userIDStr := "60d5ec49e79c9e001a8d0b1a" // Hardcoded user ID

	hc.logger.Info("Attempting to import Postman collection from URL", zap.String("url", postmanURL), zap.String("userID", userIDStr))

	// Pass c.Context() to PostmanAPIService methods if they expect it.
	// Assuming ImportCollectionByPostmanURL and GetCollection in PostmanAPIService now accept context.Context as the first argument.
	rawPostmanJSON, extractedName, err := hc.postmanAPIService.ImportCollectionByPostmanURL(c.Context(), postmanURL, collectionName)
	if err != nil {
		hc.logger.Error("Failed to import Postman collection from URL via service", zap.Error(err), zap.String("url", postmanURL))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to import from URL: " + err.Error(),
		})
	}

	finalCollectionName := collectionName
	if finalCollectionName == "" {
		finalCollectionName = extractedName
	}
	if finalCollectionName == "" { // Fallback if still no name
		finalCollectionName = "Imported from URL"
	}

	// Create the collection in the database
	createReq := models.CreateCollectionRequest{
		Name:        finalCollectionName,
		Description: "Imported from URL: " + postmanURL,
		PostmanData: rawPostmanJSON,
	}

	createdCollection, err := hc.collectionService.CreateCollection(&createReq, userIDStr) // Removed c.Context()
	if err != nil {
		hc.logger.Error("Failed to create collection after importing from URL", zap.Error(err), zap.String("url", postmanURL))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save imported collection: " + err.Error(),
		})
	}

	_, specContent, err := hc.collectionService.GenerateOpenAPISpec(createdCollection.ID.Hex()) // Removed c.Context()
	if err != nil {
		hc.logger.Error("Failed to generate OpenAPI spec after importing from URL", zap.Error(err), zap.String("collectionID", createdCollection.ID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Collection imported (ID: %s), but failed to generate OpenAPI: %s", createdCollection.ID.Hex(), err.Error()),
			"data":    fiber.Map{"collection_id": createdCollection.ID.Hex()},
		})
	}

	// Return success JSON with data for the OpenAPI preview step
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection imported and converted successfully!",
		"data": fiber.Map{
			"collection_id":   createdCollection.ID.Hex(),
			"openapi_spec":    specContent,
			"collection_name": createdCollection.Name,
		},
	})
}

// CreateCollectionFromPublicAPIHTML handles importing a public API and returns JSON
func (hc *HTMXController) CreateCollectionFromPublicAPIHTML(c fiber.Ctx) error {
	var req struct {
		PostmanID string `json:"postman_id"`
		Name      string `json:"name"`
		BaseURL   string `json:"base_url"`
	}

	if err := c.Bind().Body(&req); err != nil {
		hc.logger.Error("Invalid request body for CreateCollectionFromPublicAPIHTML", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid request: " + err.Error()})
	}

	if req.PostmanID == "" {
		hc.logger.Error("PostmanID is required for CreateCollectionFromPublicAPIHTML")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Postman ID (Collection UID) is required"})
	}

	userIDStr := "60d5ec49e79c9e001a8d0b1a" // Hardcoded user ID

	hc.logger.Info("Importing public API as collection", zap.String("postmanCollectionUID", req.PostmanID), zap.String("requestedName", req.Name), zap.String("userID", userIDStr))

	rawPostmanJSON, err := hc.postmanAPIService.GetCollection(c.Context(), req.PostmanID) // Pass c.Context()
	if err != nil {
		hc.logger.Error("Failed to fetch public API collection data using GetCollection", zap.Error(err), zap.String("postmanCollectionUID", req.PostmanID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch API data: " + err.Error()})
	}

	collectionName := req.Name
	if collectionName == "" {
		collectionName = "Public API Import - " + req.PostmanID
	}

	createReq := models.CreateCollectionRequest{
		Name:        collectionName,
		Description: "Imported from Public API list: " + req.Name + " (ID: " + req.PostmanID + ")",
		PostmanData: rawPostmanJSON,
	}

	createdCollection, err := hc.collectionService.CreateCollection(&createReq, userIDStr) // Removed c.Context()
	if err != nil {
		hc.logger.Error("Failed to create collection from public API", zap.Error(err), zap.String("postmanCollectionUID", req.PostmanID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save imported API as collection: " + err.Error()})
	}

	_, specContent, err := hc.collectionService.GenerateOpenAPISpec(createdCollection.ID.Hex()) // Removed c.Context()
	if err != nil {
		hc.logger.Error("Failed to generate OpenAPI spec from public API import", zap.Error(err), zap.String("collectionID", createdCollection.ID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("API imported as collection (ID: %s), but failed to convert to OpenAPI: %s", createdCollection.ID.Hex(), err.Error()),
			"data":    fiber.Map{"collection_id": createdCollection.ID.Hex()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Public API imported and converted successfully!",
		"data": fiber.Map{
			"collection_id":   createdCollection.ID.Hex(),
			"openapi_spec":    specContent,
			"collection_name": createdCollection.Name,
			"base_url":        req.BaseURL,
		},
	})
}

// GetGenerationStatusHTML returns HTML fragment for SDK generation status
func GetGenerationStatusHTML(c fiber.Ctx) error {
	taskID := c.Params("taskID")
	onceStatus.Do(func() {
		generationStatusTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "generation_status.html")))
	})
	c.Set("Content-Type", "text/html")
	err := generationStatusTemplate.Execute(c.Response().BodyWriter(), fiber.Map{"TaskID": taskID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// CancelGenerationTaskHTML handles cancellation of an SDK generation task
func CancelGenerationTaskHTML(c fiber.Ctx) error {
	taskID := c.Params("taskID")
	onceCancel.Do(func() {
		cancelGenerationTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "cancel_generation.html")))
	})
	c.Set("Content-Type", "text/html")
	err := cancelGenerationTemplate.Execute(c.Response().BodyWriter(), fiber.Map{"TaskID": taskID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// GetUserProfileCardHTML returns HTML fragment for the user profile card
func (hc *HTMXController) GetUserProfileCardHTML(c fiber.Ctx) error {
	data := fiber.Map{
		"IsLoggedIn": false,
		"UserID":     "",
	}
	c.Set("Content-Type", "text/html")
	err := hc.templates.ExecuteTemplate(c.Response().BodyWriter(), "user_profile_card.html", data)
	if err != nil {
		hc.logger.Error("Failed to execute template user_profile_card.html", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// GetThemeToggleHTML returns HTML fragment for the theme toggle button
func GetThemeToggleHTML(c fiber.Ctx) error {
	buttonIcon := "fa-moon"
	nextTheme := "dark"
	if c.Cookies("theme") == "dark" { // Check cookie for current theme
		buttonIcon = "fa-sun"
		nextTheme = "light"
	}
	onceTheme.Do(func() {
		themeToggleTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "theme_toggle.html")))
	})
	c.Set("Content-Type", "text/html")
	err := themeToggleTemplate.Execute(c.Response().BodyWriter(), fiber.Map{"NextTheme": nextTheme, "ButtonIcon": buttonIcon})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}

// HandleThemeToggle handles theme toggling
func HandleThemeToggle(c fiber.Ctx) error {
	theme := c.FormValue("theme")
	newButtonIcon := "fa-moon"
	nextTheme := "dark"
	cookieValue := "light"

	if theme == "dark" {
		newButtonIcon = "fa-sun"
		nextTheme = "light"
		cookieValue = "dark"
	}
	c.Cookie(&fiber.Cookie{Name: "theme", Value: cookieValue, Path: "/", Expires: time.Now().Add(365 * 24 * time.Hour)}) // Set theme cookie

	onceThemeResponse.Do(func() {
		themeToggleResponseTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "theme_toggle_response.html")))
	})

	c.Set("Content-Type", "text/html")
	err := themeToggleResponseTemplate.Execute(c.Response().BodyWriter(), fiber.Map{
		"NextTheme":     nextTheme,
		"NewButtonIcon": newButtonIcon,
		"Theme":         theme,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render template")
	}
	return nil
}
