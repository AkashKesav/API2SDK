package services

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models" // Added import for models
	"github.com/AkashKesav/API2SDK/internal/repositories"
)

// PlatformSettingsService defines the interface for platform settings management.
type PlatformSettingsService interface {
	GetPlatformSettingsMap(ctx context.Context) (map[string]interface{}, error)
	UpdatePlatformSettings(ctx context.Context, settings map[string]interface{}) (map[string]interface{}, error)
	GetPlatformSettingsStruct(ctx context.Context) (*models.PlatformSettings, error) // New method
}

// PlatformSettingsServiceImpl implements PlatformSettingsService.
type PlatformSettingsServiceImpl struct {
	repository repositories.PlatformSettingsRepository
}

// NewPlatformSettingsService creates a new PlatformSettingsServiceImpl.
func NewPlatformSettingsService(repo repositories.PlatformSettingsRepository) PlatformSettingsService {
	return &PlatformSettingsServiceImpl{repository: repo}
}

// GetPlatformSettingsMap retrieves the current platform settings as a map.
func (s *PlatformSettingsServiceImpl) GetPlatformSettingsMap(ctx context.Context) (map[string]interface{}, error) {
	settingsDoc, err := s.repository.GetSettings(ctx) // This line will now use models.PlatformSettings from the repo
	if err != nil {
		return nil, err
	}
	if settingsDoc == nil { // Handle case where GetSettings might return nil, nil
		return make(map[string]interface{}), nil // Or some other default
	}
	return settingsDoc.Settings, nil
}

// GetPlatformSettingsStruct retrieves the current platform settings as a struct.
func (s *PlatformSettingsServiceImpl) GetPlatformSettingsStruct(ctx context.Context) (*models.PlatformSettings, error) {
	settingsDoc, err := s.repository.GetSettings(ctx) // This line will now use models.PlatformSettings from the repo
	if err != nil {
		return nil, err
	}
	if settingsDoc == nil { // Handle case where GetSettings might return nil, nil
		// Return an empty struct or a default if no settings are found
		return &models.PlatformSettings{Settings: make(map[string]interface{})}, nil
	}
	return settingsDoc, nil
}

// UpdatePlatformSettings updates the platform settings.
func (s *PlatformSettingsServiceImpl) UpdatePlatformSettings(ctx context.Context, newSettings map[string]interface{}) (map[string]interface{}, error) {
	updatedSettingsDoc, err := s.repository.UpdateSettings(ctx, newSettings) // This line will now use models.PlatformSettings
	if err != nil {
		return nil, err
	}
	if updatedSettingsDoc == nil { // Handle case where UpdateSettings might return nil, nil
		return newSettings, nil // Or some other appropriate response
	}
	return updatedSettingsDoc.Settings, nil
}
