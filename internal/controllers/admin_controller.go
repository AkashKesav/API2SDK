package controllers

import (
	"bufio"
	"context" // Added for passing context to service methods
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/repositories"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AdminController handles admin-related requests
type AdminController struct {
	userService             services.UserService
	platformSettingsService services.PlatformSettingsService
	sdkService              services.SDKServiceInterface
	collectionService       *services.CollectionService // Added for active projects
	logger                  *zap.Logger
}

// NewAdminController creates a new AdminController
func NewAdminController(userService services.UserService, platformSettingsService services.PlatformSettingsService, sdkService services.SDKServiceInterface, collectionService *services.CollectionService, logger *zap.Logger) *AdminController {
	return &AdminController{
		userService:             userService,
		platformSettingsService: platformSettingsService,
		sdkService:              sdkService,
		collectionService:       collectionService,
		logger:                  logger,
	}
}

// GetStats handles GET /admin/stats
func (ac *AdminController) GetStats(c fiber.Ctx) error {
	totalUsers, err := ac.userService.GetTotalUsersCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get total users count", zap.Error(err))
		totalUsers = 0
	}

	// Get generated SDKs count
	generatedSDKs, err := ac.sdkService.GetTotalGeneratedSDKsCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get total generated SDKs count", zap.Error(err))
		generatedSDKs = 0
	}

	// Get active sessions count (from the last 24 hours)
	activeSessions := int64(0)
	// Assuming 'db' is available in the controller's scope or passed in constructor
	// For now, we'll assume it's accessible via configs.GetDatabase() as it's a global singleton.
	// A better approach would be to inject these services into the AdminController.
	sessionService := services.NewSessionService(repositories.NewSessionRepository(configs.GetDatabase()))
	if sessionService != nil {
		activeSessions, err = sessionService.GetSessionCountSince(c.Context(), time.Now().Add(-24*time.Hour))
		if err != nil {
			ac.logger.Error("Failed to get active sessions count", zap.Error(err))
			activeSessions = 0
		}
	}

	// Get API calls made today
	apiCallsToday := int64(0)
	apiLogService := services.NewAPILogService(repositories.NewAPILogRepository(configs.GetDatabase()))
	if apiLogService != nil {
		// Get the start of today
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		apiCallsToday, err = apiLogService.GetAPICallCountSince(c.Context(), startOfDay)
		if err != nil {
			ac.logger.Error("Failed to get API calls count for today", zap.Error(err))
			apiCallsToday = 0
		}
	}

	statsMap := fiber.Map{
		"total_users":     totalUsers,
		"generated_sdks":  generatedSDKs,
		"active_sessions": activeSessions,
		"api_calls_today": apiCallsToday,
	}

	return utils.SuccessResponse(c, "Successfully retrieved system statistics", statsMap)
}

// GetDashboardData handles GET /admin/dashboard
func (ac *AdminController) GetDashboardData(c fiber.Ctx) error {
	// Get total users count
	totalUsers, err := ac.userService.GetTotalUsersCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get total users count", zap.Error(err))
		totalUsers = -1
	}

	// Get platform settings to check maintenance mode
	settings, err := ac.platformSettingsService.GetPlatformSettingsStruct(c.Context())
	maintenanceMode := false
	if err == nil {
		maintenanceMode = settings.MaintenanceMode
	} else {
		ac.logger.Error("Failed to get platform settings", zap.Error(err))
	}

	// Get recent users count (users registered in the last 7 days)
	recentUsersCount, err := ac.userService.GetRecentUsersCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get recent users count", zap.Error(err))
		recentUsersCount = 0
	}

	// Get recent SDKs generated count
	recentSDKsGenerated, err := ac.sdkService.GetTotalGeneratedSDKsCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get recent SDKs generated count", zap.Error(err))
		recentSDKsGenerated = 0
	}

	// Get active projects count
	activeProjects, err := ac.collectionService.GetActiveProjectsCount(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get active projects count", zap.Error(err))
		activeProjects = 0
	}

	dashboardDataMap := fiber.Map{
		"total_users":             totalUsers,
		"maintenance_mode_active": maintenanceMode,
		"recent_users_count":      recentUsersCount,
		"recent_sdks_generated":   recentSDKsGenerated,
		"active_projects":         activeProjects,
	}

	return utils.SuccessResponse(c, "Successfully retrieved admin dashboard data", dashboardDataMap)
}

// GetAllUsers handles GET /admin/users - for listing all users
func (ac *AdminController) GetAllUsers(c fiber.Ctx) error {
	pageQuery := c.Query("page", "1")
	limitQuery := c.Query("limit", "10")

	page, err := strconv.Atoi(pageQuery)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Pass context.Background() or c.UserContext() to the service method
	users, total, err := ac.userService.GetAllUsers(context.Background(), page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get users", err.Error())
	}
	return utils.SuccessResponse(c, "Successfully retrieved users", fiber.Map{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUserByID handles GET /admin/users/:id - for fetching a specific user by admin
func (ac *AdminController) GetUserByID(c fiber.Ctx) error {
	userIDParam := c.Params("id")
	// Pass context.Background() or c.UserContext() to the service method
	user, err := ac.userService.GetUserProfile(context.Background(), userIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found or error fetching user", err.Error())
	}
	return utils.SuccessResponse(c, "Successfully retrieved user", user)
}

// UpdateUserRole handles PUT /admin/users/:id/role - for changing a user's role
func (ac *AdminController) UpdateUserRole(c fiber.Ctx) error {
	userIDParam := c.Params("id")
	var req models.UpdateUserRoleRequest
	body := c.Body()
	if len(body) == 0 {
		ac.logger.Error("Request body is empty for UpdateUserRole", zap.String("userID", userIDParam))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", "Request body is empty")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		ac.logger.Error("Invalid request body for UpdateUserRole", zap.String("userID", userIDParam), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Role != models.RoleAdmin && req.Role != models.RoleUser && req.Role != models.RoleModerator {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user role specified", "role: must be admin, user, or moderator")
	}

	// Pass context.Background() or c.UserContext() to the service method
	err := ac.userService.UpdateUserRole(context.Background(), userIDParam, req.Role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user role", err.Error())
	}
	return utils.SuccessResponse(c, "User role updated successfully", fiber.Map{
		"updated_user_id": userIDParam,
		"new_role":        req.Role,
	})
}

// DeleteUser handles DELETE /admin/users/:id - for deleting a user by admin
func (ac *AdminController) DeleteUser(c fiber.Ctx) error {
	userIDParam := c.Params("id")
	// Pass context.Background() or c.UserContext() to the service method
	err := ac.userService.DeleteUserAsAdmin(context.Background(), userIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}
	return utils.SuccessResponse(c, "User deleted successfully by admin", fiber.Map{
		"deleted_user_id": userIDParam,
	})
}

// GetPlatformSettings handles GET /admin/settings
func (ac *AdminController) GetPlatformSettings(c fiber.Ctx) error {
	settings, err := ac.platformSettingsService.GetPlatformSettingsMap(c.Context()) // Changed to GetPlatformSettingsMap and use c.Context()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve platform settings", err.Error())
	}
	return utils.SuccessResponse(c, "Successfully retrieved platform settings", settings)
}

// UpdatePlatformSettings handles PUT /admin/settings
func (ac *AdminController) UpdatePlatformSettings(c fiber.Ctx) error {
	var req models.PlatformSettingsRequest // This is now a struct with pointers
	body := c.Body()
	if len(body) == 0 {
		ac.logger.Error("Request body is empty for UpdatePlatformSettings")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body for platform settings", "Request body is empty")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		ac.logger.Error("Invalid request body for UpdatePlatformSettings", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body for platform settings", err.Error())
	}

	// req is already a map[string]interface{} due to type PlatformSettingsRequest map[string]interface{}
	// We need to validate the types of the provided fields.
	// The service expects a map[string]interface{}, so we can pass req directly after validation,
	// or construct a new validatedMap. For safety, let's construct a validatedMap.

	validatedMap := make(map[string]interface{})

	if val, ok := req["postmanApiKey"]; ok {
		if strVal, isString := val.(string); isString {
			validatedMap["postmanApiKey"] = strVal
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for postmanApiKey, must be a string", "postmanApiKey: must be a string")
		}
	}

	if val, ok := req["maintenanceMode"]; ok {
		if boolVal, isBool := val.(bool); isBool {
			validatedMap["maintenanceMode"] = boolVal
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for maintenanceMode, must be a boolean", "maintenanceMode: must be a boolean")
		}
	}

	if val, ok := req["logConfig"]; ok {
		// Assuming logConfig is expected to be a map (struct equivalent in JSON)
		if mapVal, isMap := val.(map[string]interface{}); isMap {
			// Further validation of logConfig fields can be done here if needed
			// For now, accept it as a map.
			// Example validation for a field within logConfig:
			//
			//	if enabledVal, enabledOk := mapVal["enabled"]; enabledOk {
			//	 if _, enabledIsBool := enabledVal.(bool); !enabledIsBool {
			//	     return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for logConfig.enabled, must be a boolean", "logConfig.enabled: must be a boolean")
			//	 }
			//	}
			validatedMap["logConfig"] = mapVal
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for logConfig, must be an object", "logConfig: must be an object")
		}
	}

	if val, ok := req["settings"]; ok {
		if mapVal, isMap := val.(map[string]interface{}); isMap {
			validatedMap["settings"] = mapVal
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for settings, must be an object", "settings: must be an object")
		}
	}

	if len(validatedMap) == 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "No valid settings provided for update", "")
	}

	updatedSettings, err := ac.platformSettingsService.UpdatePlatformSettings(c.Context(), validatedMap) // Use c.Context() and validatedMap
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update platform settings", err.Error())
	}
	return utils.SuccessResponse(c, "Platform settings updated successfully", updatedSettings)
}

// GetSystemLogs handles GET /admin/logs
func (ac *AdminController) GetSystemLogs(c fiber.Ctx) error {
	settings, err := ac.platformSettingsService.GetPlatformSettingsStruct(c.Context())
	if err != nil {
		ac.logger.Error("Failed to retrieve platform settings for logs", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Could not retrieve log settings", err.Error())
	}

	if !settings.LogConfig.Enabled {
		return utils.SuccessResponse(c, "Log retrieval via API is disabled.", fiber.Map{"logs": []string{}})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))          // Default 100 lines
	level := c.Query("level", settings.LogConfig.DefaultLevel) // Filter by level
	searchTerm := c.Query("search", "")                        // Basic text search

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 { // Max 1000 lines per request for safety
		limit = 100
	}

	switch settings.LogConfig.SourceType {
	case "file":
		logFilePath := settings.LogConfig.SourceDetails
		if logFilePath == "" {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Log file path not configured.", "Log file path is empty in settings")
		}

		// Sanitize the file path to prevent directory traversal
		cleanPath, err := filepath.Abs(logFilePath)
		if err != nil {
			ac.logger.Error("Failed to get absolute path for log file", zap.String("path", logFilePath), zap.Error(err))
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid log file path", err.Error())
		}
		if !strings.HasPrefix(cleanPath, "/run/media/akash/Sandisk Ultra/API2SDK/logs") {
			ac.logger.Warn("Attempt to access file outside of designated log directory", zap.String("path", cleanPath))
			return utils.ErrorResponse(c, fiber.StatusForbidden, "Access to the specified log file is not allowed.", "")
		}

		file, err := os.Open(cleanPath)
		if err != nil {
			ac.logger.Error("Failed to open log file", zap.String("path", cleanPath), zap.Error(err))
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Could not open log file", err.Error())
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Apply filters
			if level != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(level)) {
				continue
			}
			if searchTerm != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(searchTerm)) {
				continue
			}
			lines = append(lines, line)
		}

		if err := scanner.Err(); err != nil {
			ac.logger.Error("Error reading log file", zap.String("path", logFilePath), zap.Error(err))
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error reading log file", err.Error())
		}

		// Reverse lines to get latest first (simple approach for now)
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}

		// Paginate
		start := (page - 1) * limit
		end := start + limit
		if start > len(lines) {
			start = len(lines)
		}
		if end > len(lines) {
			end = len(lines)
		}
		paginatedLines := lines[start:end]

		return utils.SuccessResponse(c, "Successfully retrieved system logs", fiber.Map{
			"logs":        paginatedLines,
			"total_lines": len(lines), // Total lines matching filter before pagination
			"page":        page,
			"limit":       limit,
		})

	case "database", "external_service":
		return utils.ErrorResponse(c, fiber.StatusNotImplemented, "Log retrieval from database or external service is not yet implemented.", "Feature not available")
	default:
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Unknown log source type configured.", "Invalid log source type: "+settings.LogConfig.SourceType)
	}
}

// ToggleMaintenanceMode handles POST /admin/maintenance
func (ac *AdminController) ToggleMaintenanceMode(c fiber.Ctx) error {
	settings, err := ac.platformSettingsService.GetPlatformSettingsStruct(c.Context())
	if err != nil {
		ac.logger.Error("Failed to get platform settings for maintenance mode toggle", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Could not retrieve current settings", err.Error())
	}

	newMaintenanceModeState := !settings.MaintenanceMode

	updateMap := map[string]interface{}{
		"maintenanceMode": newMaintenanceModeState,
	}

	updatedSettings, err := ac.platformSettingsService.UpdatePlatformSettings(c.Context(), updateMap)
	if err != nil {
		ac.logger.Error("Failed to update platform settings for maintenance mode toggle", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Could not update maintenance mode", err.Error())
	}

	// The updatedSettings map from the service should contain the new state.
	// We can also directly use newMaintenanceModeState if confident about the update.
	finalState, ok := updatedSettings["maintenanceMode"].(bool)
	if !ok {
		// This case should ideally not happen if the service returns the updated field correctly.
		ac.logger.Error("Maintenance mode status missing or not a boolean in settings update response")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error confirming maintenance mode status", "Invalid settings format after update")
	}

	return utils.SuccessResponse(c, "Maintenance mode toggled successfully", fiber.Map{
		"maintenance_mode_active": finalState,
	})
}

// ManageUsers is deprecated by GetAllUsers but kept for compatibility if previously referenced directly.
// It's better to use GetAllUsers for clarity.
func (ac *AdminController) ManageUsers(c fiber.Ctx) error {
	return ac.GetAllUsers(c)
}
