package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/configs"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrMissingToken     = errors.New("missing authorization token")
	ErrInvalidToken     = errors.New("invalid token format")
	ErrExpiredToken     = errors.New("token has expired")
	ErrUnexpectedMethod = errors.New("unexpected signing method")
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// UserContext holds user information extracted from JWT
type UserContext struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// GenerateAccessToken generates a new JWT access token
func GenerateAccessToken(userID, email, role string, config *configs.Config) (string, error) {
	return generateToken(userID, email, role, "access", time.Hour*24, config) // 24 hours
}

// GenerateRefreshToken generates a new JWT refresh token
func GenerateRefreshToken(userID, email, role string, config *configs.Config) (string, error) {
	return generateToken(userID, email, role, "refresh", time.Hour*24*7, config) // 7 days
}

// generateToken is a helper function to generate JWT tokens
func generateToken(userID, email, role, tokenType string, duration time.Duration, config *configs.Config) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "api2sdk",
			Audience:  []string{"api2sdk-client"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string, config *configs.Config) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedMethod
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// JWTMiddleware validates JWT tokens and sets user context
func JWTMiddleware(config *configs.Config, logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")

		// Debug logging (reduced verbosity)
		logger.Debug("JWT Middleware processing request",
			zap.String("path", c.Path()),
			zap.Bool("hasAuthHeader", authHeader != ""),
		)

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header required",
				"error":   "missing_token",
			})
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format",
				"error":   "invalid_format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := ValidateToken(tokenString, config)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.String("error", err.Error()),
				zap.String("path", c.Path()),
			)

			var errorCode string
			var message string

			switch {
			case errors.Is(err, ErrExpiredToken):
				errorCode = "token_expired"
				message = "Token has expired"
			case errors.Is(err, ErrInvalidToken):
				errorCode = "invalid_token"
				message = "Invalid token"
			default:
				errorCode = "token_error"
				message = "Token validation failed"
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": message,
				"error":   errorCode,
			})
		}

		// Ensure it's an access token
		if claims.Type != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token type",
				"error":   "wrong_token_type",
			})
		}

		// Set user context
		userCtx := UserContext{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		}

		c.Locals("user", userCtx)
		c.Locals("user_id", claims.UserID)

		logger.Debug("JWT validation successful",
			zap.String("user_id", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("path", c.Path()),
		)

		return c.Next()
	}
}

// OptionalJWTMiddleware validates JWT tokens if present but doesn't require them
func OptionalJWTMiddleware(config *configs.Config, logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next() // No token, continue without setting user context
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next() // Invalid format, continue without setting user context
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString, config)
		if err != nil {
			logger.Debug("Optional token validation failed", zap.Error(err))
			return c.Next() // Invalid token, continue without setting user context
		}

		if claims.Type == "access" {
			userCtx := UserContext{
				ID:    claims.UserID,
				Email: claims.Email,
				Role:  claims.Role,
			}
			c.Locals("user", userCtx)
			c.Locals("user_id", claims.UserID)
		}

		return c.Next()
	}
}

// GetUser retrieves the user context from fiber context
func GetUser(c fiber.Ctx) (*UserContext, bool) {
	user := c.Locals("user")
	if user == nil {
		return nil, false
	}

	userCtx, ok := user.(UserContext)
	if !ok {
		// Debug: Log the actual type for troubleshooting
		// Note: This is a silent failure - we don't log here to avoid spam
		return nil, false
	}

	return &userCtx, true
}

// GetUserID retrieves the user ID from fiber context
func GetUserID(c fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// RequireRole creates a middleware that checks if user has required role
func RequireRole(requiredRole string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, exists := GetUser(c)
		if !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
				"error":   "no_user_context",
			})
		}

		if user.Role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success":       false,
				"message":       "Insufficient permissions",
				"error":         "role_required",
				"required_role": requiredRole,
				"user_role":     user.Role,
			})
		}

		return c.Next()
	}
}

// RequireAnyRole creates a middleware that checks if user has any of the required roles
func RequireAnyRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, exists := GetUser(c)
		if !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
				"error":   "no_user_context",
			})
		}

		for _, role := range roles {
			if user.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success":        false,
			"message":        "Insufficient permissions",
			"error":          "role_required",
			"required_roles": roles,
			"user_role":      user.Role,
		})
	}
}

// RefreshTokenMiddleware validates refresh tokens specifically
func RefreshTokenMiddleware(config *configs.Config, logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header required",
				"error":   "missing_token",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format",
				"error":   "invalid_format",
			})
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString, config)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid refresh token",
				"error":   "invalid_token",
			})
		}

		// Ensure it's a refresh token
		if claims.Type != "refresh" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token type - refresh token required",
				"error":   "wrong_token_type",
			})
		}

		userCtx := UserContext{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		}

		c.Locals("user", userCtx)
		c.Locals("user_id", claims.UserID)

		return c.Next()
	}
}
