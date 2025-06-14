package models

// UpdateUserProfileRequest defines the expected structure for profile update requests.
type UpdateUserProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	// Add other fields that can be updated, e.g., Bio, AvatarURL, etc.
}

// ChangePasswordRequest defines the structure for password change requests.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// UpdateUserRoleRequest defines the structure for updating a user's role.
type UpdateUserRoleRequest struct {
	Role UserRole `json:"role" validate:"required"`
}

// PlatformSettingsRequest defines the structure for updating platform settings.
// It's a map to allow flexible settings.
type PlatformSettingsRequest map[string]interface{}

// MaintenanceModeRequest defines the structure for toggling maintenance mode.
type MaintenanceModeRequest struct {
	Enable bool `json:"enable"`
}
