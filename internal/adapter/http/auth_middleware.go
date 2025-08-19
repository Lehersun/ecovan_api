package http

import (
	"context"
	"net/http"
	"strings"

	"eco-van-api/internal/adapter/auth"
	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

// AuthMiddleware provides authentication for protected routes
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth middleware that requires a valid access token
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteUnauthorized(w, "Authorization header is required")
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			WriteUnauthorized(w, "Invalid authorization header format")
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			WriteUnauthorized(w, "Token is required")
			return
		}

		// Validate the access token
		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			WriteUnauthorized(w, "Invalid or expired token")
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware that requires a specific user role
func (m *AuthMiddleware) RequireRole(requiredRole models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, ensure user is authenticated
			userRole, ok := r.Context().Value(userRoleKey).(models.UserRole)
			if !ok {
				WriteUnauthorized(w, "User role not found in context")
				return
			}

			// Check if user has the required role
			if userRole != requiredRole {
				WriteForbidden(w, "Insufficient permissions")
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, ensure user is authenticated
			userRole, ok := r.Context().Value(userRoleKey).(models.UserRole)
			if !ok {
				WriteUnauthorized(w, "User role not found in context")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range requiredRoles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				WriteForbidden(w, "Insufficient permissions")
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserRoleFromContext extracts the user role from the request context
func GetUserRoleFromContext(ctx context.Context) (models.UserRole, bool) {
	userRole, ok := ctx.Value(userRoleKey).(models.UserRole)
	return userRole, ok
}
