package controllers

import (
	"bufio"
	"context" // Added for passing context to service methods
	"os"
	"strconv"
	"strings"

	"github.com/AkashKesav/API2SDK/internal/middleware"
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
	logger                  *zap.Logger // Added logger
}

// NewAdminController creates a new AdminController
func NewAdminController(userService services.UserService, platformSettingsService services.PlatformSettingsService, logger *zap.Logger) *AdminController { // Added logger
	return &AdminController{
		userService:             userService,
		platformSettingsService: platformSettingsService,
		logger:                  logger, // Store logger
	}
}

// GetStats handles GET /admin/stats
// TODO: Implement actual system statistics retrieval
func (ac *AdminController) GetStats(c fiber.Ctx) error {
	totalUsers, err := ac.userService.GetTotalUsersCount(c.Context())
	if err != nil {
		// Log the error but don't fail the entire request, return 0 for users
		// Or handle more gracefully depending on requirements
		totalUsers = 0
		// Consider logging: ac.logger.Error("Failed to get total users count", zap.Error(err))
	}

	// TODO: Implement service calls for these stats
	// These would require new methods in relevant services (e.g., SdkService, some session service)
	// and corresponding repository methods.
	generatedSDKs := int64(0)  // Placeholder - Assuming SdkService.GetTotalGeneratedSDKsCount()
	activeSessions := int64(0) // Placeholder - Assuming some SessionService.GetActiveSessionsCount()
	apiCallsToday := int64(0)  // Placeholder - Assuming some APILogService.GetAPICallsCount(today)

	userID, userIDOk := middleware.GetUserID(c)
	userRole, userRoleOk := middleware.GetUserID(c) // This was a typo, should be GetUserRole

	statsMap := fiber.Map{
		"total_users":     totalUsers,
		"generated_sdks":  generatedSDKs,
		"active_sessions": activeSessions,
		"api_calls_today": apiCallsToday,
	}
	if userIDOk {
		statsMap["user_id"] = userID // Admin making the request
	}
	if userRoleOk {
		statsMap["role"] = userRole // Admin's role
	}

	return utils.SuccessResponse(c, "Successfully retrieved system statistics", statsMap)
}

// GetDashboardData handles GET /admin/dashboard
// TODO: Implement actual dashboard data retrieval
func (ac *AdminController) GetDashboardData(c fiber.Ctx) error {
	totalUsers, err := ac.userService.GetTotalUsersCount(c.Context())
	if err != nil {
		// Log error, but continue, providing a partial dashboard
		// ac.logger.Error("Failed to get total users for dashboard", zap.Error(err))
		totalUsers = -1 // Indicate an error or unavailable data
	}

	settings, err := ac.platformSettingsService.GetPlatformSettingsStruct(c.Context())
	maintenanceMode := false // Default to false if settings can't be fetched
	if err == nil {
		maintenanceMode = settings.MaintenanceMode
	} else {
		// Log error
		// ac.logger.Error("Failed to get platform settings for dashboard", zap.Error(err))
	}

	// Placeholders for more detailed stats that would require dedicated service methods
	recentUsersCount := 0    // e.g., users registered in the last 7 days
	recentSDKsGenerated := 0 // e.g., SDKs generated in the last 7 days
	activeProjects := 0      // Placeholder for a count of active/used collections or APIs

	userID, userIDOk := middleware.GetUserID(c)
	user, userOk := middleware.GetUser(c)
	userRole := ""
	userRoleOk := false
	if userOk {
		userRole = user.Role
		userRoleOk = true
	}

	dashboardDataMap := fiber.Map{
		"total_users":             totalUsers,
		"maintenance_mode_active": maintenanceMode,
		"recent_users_count":      recentUsersCount,    // Placeholder
		"recent_sdks_generated":   recentSDKsGenerated, // Placeholder
		"active_projects":         activeProjects,      // Placeholder
	}
	if userIDOk {
		dashboardDataMap["user_id"] = userID
	}
	if userRoleOk {
		dashboardDataMap["role"] = userRole
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
	if err := c.Bind().Body(&req); err != nil {
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
	if err := c.Bind().Body(&req); err != nil {
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

	if val, ok := req["jwtSecretKey"]; ok {
		if strVal, isString := val.(string); isString {
			validatedMap["jwtSecretKey"] = strVal
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type for jwtSecretKey, must be a string", "jwtSecretKey: must be a string")
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

		file, err := os.Open(logFilePath)
		if err != nil {
			ac.logger.Error("Failed to open log file", zap.String("path", logFilePath), zap.Error(err))
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
