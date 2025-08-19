package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"eco-van-api/internal/models"
)

func TestRBACMiddleware_RequireAdminRole(t *testing.T) {
	middleware := NewRBACMiddleware()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin role should pass",
			userRole:       models.UserRoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Dispatcher role should be forbidden",
			userRole:       models.UserRoleDispatcher,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Driver role should be forbidden",
			userRole:       models.UserRoleDriver,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Viewer role should be forbidden",
			userRole:       models.UserRoleViewer,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			ctx := context.WithValue(req.Context(), userRoleKey, tt.userRole)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.RequireAdminRole(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestRBACMiddleware_RequireReadAccess(t *testing.T) {
	middleware := NewRBACMiddleware()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin role should pass",
			userRole:       models.UserRoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Dispatcher role should pass",
			userRole:       models.UserRoleDispatcher,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Driver role should pass",
			userRole:       models.UserRoleDriver,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Viewer role should pass",
			userRole:       models.UserRoleViewer,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			ctx := context.WithValue(req.Context(), userRoleKey, tt.userRole)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.RequireReadAccess(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestRBACMiddleware_RequireWriteAccess(t *testing.T) {
	middleware := NewRBACMiddleware()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin role should pass",
			userRole:       models.UserRoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Dispatcher role should be forbidden",
			userRole:       models.UserRoleDispatcher,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Driver role should be forbidden",
			userRole:       models.UserRoleDriver,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Viewer role should be forbidden",
			userRole:       models.UserRoleViewer,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", http.NoBody)
			ctx := context.WithValue(req.Context(), userRoleKey, tt.userRole)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.RequireWriteAccess(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestRBACMiddleware_NoRoleInContext(t *testing.T) {
	middleware := NewRBACMiddleware()

	tests := []struct {
		name           string
		middlewareFunc func(http.Handler) http.Handler
	}{
		{
			name:           "RequireAdminRole without role should be unauthorized",
			middlewareFunc: middleware.RequireAdminRole,
		},
		{
			name:           "RequireWriteAccess without role should be unauthorized",
			middlewareFunc: middleware.RequireWriteAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			tt.middlewareFunc(handler).ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
			}
		})
	}
}
