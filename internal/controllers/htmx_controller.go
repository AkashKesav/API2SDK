package controllers

import (
	"fmt"
	"strconv"
	"time" // For cookie expiration in HandleThemeToggle

	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type HTMXController struct {
	logger            *zap.Logger
	sdkService        *services.SDKService
	sdkRepo           *repositories.SDKRepository
	collectionService *services.CollectionService
	postmanAPIService *services.PostmanAPIService
	publicAPIService  *services.PublicAPIService // Added PublicAPIService
}

func NewHTMXController(logger *zap.Logger, sdkService *services.SDKService, sdkRepo *repositories.SDKRepository, collectionService *services.CollectionService, postmanAPIService *services.PostmanAPIService, publicAPIService *services.PublicAPIService) *HTMXController {
	return &HTMXController{
		logger:            logger,
		sdkService:        sdkService,
		sdkRepo:           sdkRepo,
		collectionService: collectionService,
		postmanAPIService: postmanAPIService,
		publicAPIService:  publicAPIService, // Store injected PublicAPIService
	}
}

// GetSDKHistoryHTML returns HTML fragment for SDK history
func (hc *HTMXController) GetSDKHistoryHTML(c fiber.Ctx) error {
	hc.logger.Info("GetSDKHistoryHTML called (currently placeholder)")
	return c.SendString(`
	<div class="history-placeholder">
	<i class="fas fa-tools"></i>
	<p>SDK History feature is currently under maintenance.</p>
	</div>
	`)
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
	return c.SendString(fmt.Sprintf(`
<div class="alert alert-warning" role="alert">
  <i class="fas fa-exclamation-triangle"></i>
  Deletion of SDK %s is a placeholder and not yet implemented.
</div>`, id))
}

// GetPopularAPIsHTML returns HTML fragment for popular APIs
func (hc *HTMXController) GetPopularAPIsHTML(c fiber.Ctx) error {
	// Use the injected publicAPIService
	apis := hc.publicAPIService.GetPopularAPIs()

	if len(apis) == 0 {
		return c.SendString(`
<div class="api-placeholder">
	<i class="fas fa-search"></i>
	<p>No APIs found. Try different search terms.</p>
</div>
`)
	}

	html := ""
	for _, api := range apis {
		tagsHTML := ""
		for _, tag := range api.Tags {
			tagsHTML += fmt.Sprintf(`<span class="tag">%s</span>`, tag)
		}

		html += fmt.Sprintf(`
<div class="api-card" data-api-id="%s">
	<div class="api-header">
		<h4>%s</h4>
		<span class="api-category">%s</span>
	</div>
	<p class="api-description">%s</p>
	<div class="api-tags">
		%s
	</div>
	<div class="api-actions">
		<button class="btn btn-primary btn-sm" 
				onclick="selectAPI('%s', '%s', '%s', '%s')">
			<i class="fas fa-check"></i> Select API
		</button>
		<a href="%s" target="_blank" class="btn btn-secondary btn-sm">
			<i class="fas fa-external-link-alt"></i> View Collection
		</a>
	</div>
</div>
`, api.PostmanID, api.Name, api.Category, api.Description,
			tagsHTML, api.PostmanID, api.Name, api.BaseURL, api.PostmanURL, api.PostmanURL)
	}

	return c.SendString(html)
}

// CreateCollectionHTML handles collection creation from HTMX forms and returns HTML
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
			return c.SendString(fmt.Sprintf(`<div class=\"alert alert-error\"><i class=\"fas fa-exclamation-circle\"></i><strong>Error:</strong> Failed to read uploaded file: %s</div>`, err.Error()))
		}
		defer fileContent.Close()

		// Read file as string
		fileBytes := make([]byte, file.Size)
		_, err = fileContent.Read(fileBytes)
		if err != nil {
			return c.SendString(fmt.Sprintf(`<div class=\"alert alert-error\"><i class=\"fas fa-exclamation-circle\"></i><strong>Error:</strong> Failed to read file content: %s</div>`, err.Error()))
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
		return c.SendString(`<div class=\"alert alert-error\"><i class=\"fas fa-exclamation-circle\"></i><strong>Error:</strong> Postman collection data is required (either upload a file or paste JSON).</div>`)
	}

	// Get user ID from middleware context
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		hc.logger.Warn("CreateCollectionHTML: UserID not found or invalid in context")
		return c.Status(fiber.StatusUnauthorized).SendString(`<div class=\"alert alert-error\"><i class=\"fas fa-exclamation-circle\"></i><strong>Error:</strong> Unauthorized. Please log in.</div>`)
	}

	// Use the injected collectionService
	collection, err := hc.collectionService.CreateCollection(&req, userIDStr) // Removed c.Context()
	if err != nil {
		return c.SendString(fmt.Sprintf(`<div class="alert alert-error"><i class="fas fa-exclamation-circle"></i><strong>Error:</strong> Failed to create collection: %s</div>`, err.Error()))
	}

	// Return success HTML with JavaScript to proceed to configuration
	// Instead of directly going to config, now we go to OpenAPI preview if PostmanData was provided
	if req.PostmanData != "" {
		// If Postman data was part of the request, generate OpenAPI spec
		// The collection.ID will be available here.
		// We need to call the GenerateOpenAPISpec method from collectionSvc
		_, specContent, err := hc.collectionService.GenerateOpenAPISpec(collection.ID.Hex()) // Removed c.Context()
		if err != nil {
			hc.logger.Error("Failed to generate OpenAPI spec after collection creation", zap.Error(err), zap.String("collectionID", collection.ID.Hex()))
			// Return an error message, but the collection is created.
			// User might need to manually trigger generation or we could guide them.
			return c.SendString(fmt.Sprintf(`
			<div class="alert alert-warning">
				<i class="fas fa-exclamation-triangle"></i>
				<strong>Collection created (ID: %s), but failed to auto-generate OpenAPI spec:</strong> %s
				<p>You can try generating it manually from the collection settings.</p>
			</div>
			<script>
				// Still hide source and show preview, but with a warning
				document.getElementById('openapi-collection-id').value = '%s';
				document.getElementById('openapi-spec-preview').value = "Error generating OpenAPI spec: %s";
				document.getElementById('step-source').style.display = 'none';
				document.getElementById('upload-panel').style.display = 'none';
				document.getElementById('url-import-panel').style.display = 'none';
        document.getElementById('public-api-panel').style.display = 'none';
				document.getElementById('step-openapi-preview').style.display = 'block';
			</script>
			`, collection.ID.Hex(), err.Error(), collection.ID.Hex(), err.Error()))
		}

		// Return HTML to show OpenAPI preview step
		return c.SendString(fmt.Sprintf(`
		<div class="alert alert-success">
			<i class="fas fa-check-circle"></i>
			<strong>Success:</strong> Collection uploaded and OpenAPI spec generated! Review below.
		</div>
		<script>
			document.getElementById('openapi-collection-id').value = '%s';
			document.getElementById('openapi-spec-preview').value = %s;
			document.getElementById('step-source').style.display = 'none';
			document.getElementById('upload-panel').style.display = 'none';
			document.getElementById('url-import-panel').style.display = 'none';
      document.getElementById('public-api-panel').style.display = 'none';
			document.getElementById('step-openapi-preview').style.display = 'block';
		</script>
		`, collection.ID.Hex(), strconv.Quote(specContent)))
	} else {
		// This case should ideally not happen if PostmanData is required.
		// If it can happen (e.g. creating an empty collection to be populated later),
		// then proceed to config or a different step.
		return c.SendString(fmt.Sprintf(`
		<div class="alert alert-success">
			<i class="fas fa-check-circle"></i>
			<strong>Success:</strong> Collection created successfully (ID: %s)! No Postman data to convert.
		</div>
		<script>
			document.getElementById('collection-id').value = '%s'; // This ID is for the SDK config step
			document.getElementById('step-source').style.display = 'none';
			document.getElementById('upload-panel').style.display = 'none';
			document.getElementById('url-import-panel').style.display = 'none';
      document.getElementById('public-api-panel').style.display = 'none';
			document.getElementById('step-config').style.display = 'block'; // Go to config if no preview
		</script>
		`, collection.ID.Hex(), collection.ID.Hex()))
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
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		hc.logger.Warn("CreateCollectionFromURLHTML: UserID not found or invalid in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: User ID not found. Please log in.",
		})
	}

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

	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		hc.logger.Warn("CreateCollectionFromPublicAPIHTML: UserID not found or invalid in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: User ID not found. Please log in.",
		})
	}

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
	return c.SendString(fmt.Sprintf(
		`<div id="generation-status-%s" class="generation-status-polling">
			<p><i class="fas fa-spinner fa-spin"></i> Checking status for task %s...</p>
			<div hx-get="/api/htmx/generation-status/%s" hx-trigger="every 5s" hx-swap="outerHTML"></div>
		</div>`,
		taskID, taskID, taskID,
	))
}

// CancelGenerationTaskHTML handles cancellation of an SDK generation task
func CancelGenerationTaskHTML(c fiber.Ctx) error {
	taskID := c.Params("taskID")
	return c.SendString(fmt.Sprintf(
		`<div class="alert alert-info">
			<i class="fas fa-info-circle"></i> Cancellation requested for task %s. (Placeholder)
		</div>`,
		taskID,
	))
}

// GetUserProfileCardHTML returns HTML fragment for the user profile card
func (hc *HTMXController) GetUserProfileCardHTML(c fiber.Ctx) error {
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		return c.SendString("<p>Not logged in. <a href='/login'>Login</a></p>")
	}
	// In a real application, fetch user details
	// For placeholder, use userIDStr directly
	return c.SendString(fmt.Sprintf(
		`<div class=\"user-profile-card\">
<p><i class=\"fas fa-user\"></i> User ID: %s (Placeholder)</p>
<p><a href=\"/logout\">Logout</a></p>
</div>`,
		userIDStr,
	))
}

// GetThemeToggleHTML returns HTML fragment for the theme toggle button
func GetThemeToggleHTML(c fiber.Ctx) error {
	buttonIcon := "fa-moon"
	nextTheme := "dark"
	if c.Cookies("theme") == "dark" { // Check cookie for current theme
		buttonIcon = "fa-sun"
		nextTheme = "light"
	}
	return c.SendString(fmt.Sprintf(
		`<button hx-post="/api/htmx/theme-toggle" hx-vals='{"theme": "%s"}' hx-swap="outerHTML" class="btn btn-theme-toggle">
			<i class="fas %s"></i> Toggle Theme
		</button>`,
		nextTheme, buttonIcon,
	))
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

	return c.SendString(fmt.Sprintf(
		`<button hx-post="/api/htmx/theme-toggle" hx-vals='{"theme": "%s"}' hx-swap="outerHTML" class="btn btn-theme-toggle">
			<i class="fas %s"></i> Toggle Theme (Now %s)
		</button>`,
		nextTheme, newButtonIcon, theme,
	))
}
