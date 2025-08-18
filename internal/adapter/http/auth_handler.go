package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-van-api/internal/models"
	"eco-van-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login requests
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Attempt login
	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		WriteUnauthorized(w, "Invalid credentials")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		WriteInternalError(w, "Failed to encode response")
		return
	}
}

// Refresh handles token refresh requests
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Attempt refresh
	response, err := h.authService.Refresh(r.Context(), &req)
	if err != nil {
		WriteUnauthorized(w, "Invalid refresh token")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		WriteInternalError(w, "Failed to encode response")
		return
	}
}

// CreateUser handles user creation requests
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Create user
	user, err := h.authService.CreateUser(r.Context(), &req)
	if err != nil {
		WriteConflict(w, err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		WriteInternalError(w, "Failed to encode response")
		return
	}
}

// GetUser handles user retrieval requests
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		WriteBadRequest(w, "User ID is required")
		return
	}

	// Get user
	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		WriteNotFound(w, "User not found")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		WriteInternalError(w, "Failed to encode response")
		return
	}
}

// ListUsers handles user listing requests
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get users
	users, total, err := h.authService.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		WriteInternalError(w, "Failed to retrieve users")
		return
	}

	// Return success response with proper format
	response := map[string]interface{}{
		"total": total,
		"page":  page,
		"pageSize": pageSize,
		"items": users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		WriteInternalError(w, "Failed to encode response")
		return
	}
}

// DeleteUser handles user deletion requests
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		WriteBadRequest(w, "User ID is required")
		return
	}

	// Delete user
	err := h.authService.DeleteUser(r.Context(), userID)
	if err != nil {
		WriteNotFound(w, "User not found")
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}
