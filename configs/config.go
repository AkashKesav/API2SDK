package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server Configuration
	Port string `json:"port"`

	// MongoDB Configuration
	MongoDBURI      string `json:"mongodb_uri"`
	MongoDBName     string `json:"mongodb_name"`
	MongoDBUsername string `json:"mongodb_username"`
	MongoDBPassword string `json:"mongodb_password"`

	// Authentication Configuration
	JWTSecret string `json:"jwt_secret"`

	// External API Configuration
	PostmanAPIKey string `json:"postman_api_key"`

	// Environment
	Environment string `json:"environment"`
}

// GlobalConfig holds the global configuration instance
var GlobalConfig *Config

// InitConfig initializes the global configuration instance
// This function is used for backward compatibility with existing code
func InitConfig(config *Config) {
	GlobalConfig = config
}

// GetPostmanAPIKey returns the Postman API key from the global configuration
func GetPostmanAPIKey() string {
	if GlobalConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return GlobalConfig.PostmanAPIKey
}

// GetJWTSecret returns the JWT secret from the global configuration
func GetJWTSecret() string {
	if GlobalConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return GlobalConfig.JWTSecret
}

// GetPort returns the server port from the global configuration
func GetPort() string {
	if GlobalConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return GlobalConfig.Port
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		// Continue execution as environment variables might be set directly
	}

	config := &Config{
		// Server Configuration
		Port: getEnvOrDefault("API_PORT", "3000"),

		// MongoDB Configuration
		MongoDBURI:      getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:     getEnvOrDefault("MONGODB_DATABASE", "API2SDK"),
		MongoDBUsername: getEnvOrDefault("MONGODB_USERNAME", ""),
		MongoDBPassword: getEnvOrDefault("MONGODB_PASSWORD", ""),

		// Authentication Configuration
		JWTSecret: getEnvOrDefault("JWT_SECRET", ""),

		// External API Configuration
		PostmanAPIKey: getEnvOrDefault("POSTMAN_API_KEY", ""),

		// Environment
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set global config
	GlobalConfig = config

	log.Println("Configuration loaded successfully")
	return config, nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if GlobalConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return GlobalConfig
}

// validateConfig validates the required configuration fields
func validateConfig(config *Config) error {
	if config.MongoDBURI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}

	if config.MongoDBName == "" {
		return fmt.Errorf("MONGODB_DATABASE is required")
	}

	if config.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if config.Port == "" {
		return fmt.Errorf("API_PORT is required")
	}

	// Validate port is a valid number
	if _, err := strconv.Atoi(config.Port); err != nil {
		return fmt.Errorf("API_PORT must be a valid number: %w", err)
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetMongoDBConnectionString returns the full MongoDB connection string
func (c *Config) GetMongoDBConnectionString() string {
	return c.MongoDBURI
}

// GetServerAddress returns the server address in the format :port
func (c *Config) GetServerAddress() string {
	return ":" + c.Port
}

// LogConfig logs the current configuration (excluding sensitive data)
func (c *Config) LogConfig() {
	log.Printf("Configuration:")
	log.Printf("  Environment: %s", c.Environment)
	log.Printf("  Port: %s", c.Port)
	log.Printf("  MongoDB Database: %s", c.MongoDBName)
	log.Printf("  MongoDB URI: %s", maskSensitiveData(c.MongoDBURI))
	log.Printf("  JWT Secret: %s", maskSensitiveData(c.JWTSecret))
	log.Printf("  Postman API Key: %s", maskSensitiveData(c.PostmanAPIKey))
}

// maskSensitiveData masks sensitive configuration data for logging
func maskSensitiveData(data string) string {
	if len(data) <= 8 {
		return "***"
	}
	return data[:4] + "***" + data[len(data)-4:]
}
