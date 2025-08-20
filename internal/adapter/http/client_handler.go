package http

import (
	"net/http"
	"strconv"
	"strings"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// clientHandler handles HTTP requests for client operations
type clientHandler struct {
	clientService port.ClientService
	validate      *validator.Validate
}

// NewClientHandler creates a new client handler
func NewClientHandler(clientService port.ClientService) *clientHandler {
	return &clientHandler{
		clientService: clientService,
		validate:      validator.New(),
	}
}

// RegisterRoutes registers the client routes with the router
func (h *clientHandler) RegisterRoutes(router chi.Router) {
	router.Get("/clients", h.ListClients)
	router.Post("/clients", h.CreateClient)
	router.Get("/clients/{id}", h.GetClient)
	router.Put("/clients/{id}", h.UpdateClient)
	router.Delete("/clients/{id}", h.DeleteClient)
	router.Post("/clients/{id}/restore", h.RestoreClient)
}

// ListClients handles GET /v1/clients with pagination and filtering
func (h *clientHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	includeDeleted := r.URL.Query().Get("includeDeleted") == QueryParamIncludeDeleted

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// Create request
	req := models.ClientListRequest{
		Page:           page,
		PageSize:       pageSize,
		Query:          query,
		IncludeDeleted: includeDeleted,
	}

	// Get clients from service
	response, err := h.clientService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list clients")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// CreateClient handles POST /v1/clients
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *clientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateClientRequest

	// Parse request body
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Create client via service
	response, err := h.clientService.Create(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Client creation failed")
			return
		}
		WriteInternalError(w, "Failed to create client")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, response)
}

// GetClient handles GET /v1/clients/{id}
func (h *clientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID")
		return
	}

	// Get client from service
	response, err := h.clientService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Client not found")
			return
		}
		WriteInternalError(w, "Failed to get client")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// UpdateClient handles PUT /v1/clients/{id}
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *clientHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID")
		return
	}

	var req models.UpdateClientRequest

	// Parse request body
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Update client via service
	response, err := h.clientService.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Client not found")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Client update failed")
			return
		}
		WriteInternalError(w, "Failed to update client")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// DeleteClient handles DELETE /v1/clients/{id}
func (h *clientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID")
		return
	}

	// Delete client via service
	err = h.clientService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Client not found")
			return
		}
		WriteInternalError(w, "Failed to delete client")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RestoreClient handles POST /v1/clients/{id}/restore
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *clientHandler) RestoreClient(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID")
		return
	}

	// Restore client via service
	response, err := h.clientService.Restore(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Client not found")
			return
		}
		if strings.Contains(err.Error(), "not deleted") {
			WriteBadRequest(w, "Client is not deleted")
			return
		}
		if strings.Contains(err.Error(), "already taken") {
			WriteConflict(w, "Cannot restore client")
			return
		}
		WriteInternalError(w, "Failed to restore client")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}
