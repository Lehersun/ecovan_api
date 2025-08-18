package http

import (
	"net/http"

	"eco-van-api/internal/models"
)

// RBACMiddleware provides role-based access control for user endpoints
type RBACMiddleware struct{}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

// RequireAdminRole middleware that requires ADMIN role for full access
func (m *RBACMiddleware) RequireAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := GetUserRoleFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, "User role not found in context")
			return
		}

		if userRole != models.UserRoleAdmin {
			WriteForbidden(w, "Admin role required for this operation")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireReadAccess middleware that allows read access for all authenticated users
func (m *RBACMiddleware) RequireReadAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All authenticated users can read (ADMIN, DISPATCHER, DRIVER, VIEWER)
		// This middleware just ensures the user is authenticated (handled by RequireAuth)
		next.ServeHTTP(w, r)
	})
}

// RequireWriteAccess middleware that requires ADMIN role for write operations
func (m *RBACMiddleware) RequireWriteAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := GetUserRoleFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, "User role not found in context")
			return
		}

		// Only ADMIN can perform write operations
		if userRole != models.UserRoleAdmin {
			WriteForbidden(w, "Admin role required for write operations")
			return
		}

		next.ServeHTTP(w, r)
	})
}
