package infrastructure

import (
	"fmt"
	"net/http"
	"strings"
	domain "task_manager/Domain"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService domain.JWTService
}

func NewAuthMiddleware(jwtService domain.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

// Validates the JWT token
func (middleware *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get to fron the authrorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required."})
			return
		}

		// Check if the header is in the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be 'Bearer <token>'"})
			return
		}

		tokenString := parts[1] // Extract the token string

		// Validate the token
		claims, err := middleware.jwtService.ValidateToken(tokenString)
		if err != nil {
			// Handle invalid or expired tokens
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid or expired token: %v", err.Error())})
			return
		}

		// The token is valid. Store the user's claims in the context for later use in the handlers
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// Proceed to the next handler/middleware
		c.Next()
	}
}

// This checks if the authenicated user has one of the required roles
// It assumes AuthRequired middleware has already run and set the "role" in the context
func (middleware *AuthMiddleware) AuthorizeRole(requireRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the role from the context (set by AuthRequired middleware)
		role, exists := c.Get("role")
		if !exists {
			// This indicates a middleware chain setup error (AuthorizeRole called before AuthRequired)
			// Or AuthRequired failed but didn;t abort
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context. Authentication middleware missing?"})
			return
		}

		userRole, ok := role.(string)
		if !ok {
			// Context value is not a string? This indicaates an issue with setting the context
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format in context"})
			return
		}

		// Check if user's role is in the list of required roles
		isAuthorized := false
		for _, requiredRole := range requireRoles {
			if userRole == requiredRole {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		// User has the required role. Proceed.
		c.Next()
	}
}

// Helper function to get user details form context in a handler
func GetUserFromContext(c *gin.Context) (string, string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", "", fmt.Errorf("username is not found in context")
	}

	role, exists := c.Get("role")
	if !exists {
		return "", "", fmt.Errorf("role not found in context")
	}

	userStr, ok := username.(string)
	if !ok {
		return "", "", fmt.Errorf("username in context is not a string")
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", "", fmt.Errorf("role in context is not a string")
	}

	return userStr, roleStr, nil
}
