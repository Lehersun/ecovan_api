package http

import (
	"net/http"
	"strconv"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type clientObjectHandler struct {
	clientObjectService port.ClientObjectService
	validate            *validator.Validate
}

// NewClientObjectHandler creates a new client object HTTP handler
func NewClientObjectHandler(clientObjectService port.ClientObjectService) *clientObjectHandler {
	return &clientObjectHandler{
		clientObjectService: clientObjectService,
		validate:            validator.New(),
	}
}

// RegisterRoutes registers the client object routes with the router
func (h *clientObjectHandler) RegisterRoutes(router chi.Router) {
	router.Get("/clients/{clientId}/objects", h.ListClientObjects)
	router.Post("/clients/{clientId}/objects", h.CreateClientObject)
	router.Get("/clients/{clientId}/objects/{id}", h.GetClientObject)
	router.Put("/clients/{clientId}/objects/{id}", h.UpdateClientObject)
	router.Delete("/clients/{clientId}/objects/{id}", h.DeleteClientObject)
	router.Post("/clients/{clientId}/objects/{id}/restore", h.RestoreClientObject)
}

// ListClientObjects handles GET /v1/clients/{clientId}/objects
func (h *clientObjectHandler) ListClientObjects(w http.ResponseWriter, r *http.Request) {
	// Parse client ID from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	req := models.ClientObjectListRequest{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	}

	// Get client objects
	response, err := h.clientObjectService.List(r.Context(), clientID, req)
	if err != nil {
		WriteInternalError(w, "Failed to list client objects")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

// CreateClientObject handles POST /v1/clients/{clientId}/objects
func (h *clientObjectHandler) CreateClientObject(w http.ResponseWriter, r *http.Request) {
	// Parse client ID from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	// Parse request body
	var req models.CreateClientObjectRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Create client object
	response, err := h.clientObjectService.Create(r.Context(), clientID, req)
	if err != nil {
		if err.Error() == "client not found" {
			WriteNotFound(w, "Client not found")
			return
		}
		if err.Error()[:len("client object with name")] == "client object with name" {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to create client object")
		return
	}

	WriteJSON(w, http.StatusCreated, response)
}

// GetClientObject handles GET /v1/clients/{clientId}/objects/{id}
func (h *clientObjectHandler) GetClientObject(w http.ResponseWriter, r *http.Request) {
	// Parse IDs from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client object ID format")
		return
	}

	// Check if includeDeleted is requested
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Get client object
	response, err := h.clientObjectService.GetByID(r.Context(), clientID, id, includeDeleted)
	if err != nil {
		if err.Error() == "client not found" {
			WriteNotFound(w, "Client not found")
			return
		}
		if err.Error() == "client object not found" {
			WriteNotFound(w, "Client object not found")
			return
		}
		WriteInternalError(w, "Failed to get client object")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

// UpdateClientObject handles PUT /v1/clients/{clientId}/objects/{id}
func (h *clientObjectHandler) UpdateClientObject(w http.ResponseWriter, r *http.Request) {
	// Parse IDs from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client object ID format")
		return
	}

	// Parse request body
	var req models.UpdateClientObjectRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Update client object
	response, err := h.clientObjectService.Update(r.Context(), clientID, id, req)
	if err != nil {
		if err.Error() == "client not found" {
			WriteNotFound(w, "Client not found")
			return
		}
		if err.Error() == "client object not found" {
			WriteNotFound(w, "Client object not found")
			return
		}
		if err.Error()[:len("client object with name")] == "client object with name" {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to update client object")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

// DeleteClientObject handles DELETE /v1/clients/{clientId}/objects/{id}
func (h *clientObjectHandler) DeleteClientObject(w http.ResponseWriter, r *http.Request) {
	// Parse IDs from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client object ID format")
		return
	}

	// Delete client object
	err = h.clientObjectService.Delete(r.Context(), clientID, id)
	if err != nil {
		if err.Error() == "client not found" {
			WriteNotFound(w, "Client not found")
			return
		}
		if err.Error() == "client object not found" {
			WriteNotFound(w, "Client object not found")
			return
		}
		if err.Error()[:len("cannot delete client object")] == "cannot delete client object" {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to delete client object")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreClientObject handles POST /v1/clients/{clientId}/objects/{id}/restore
func (h *clientObjectHandler) RestoreClientObject(w http.ResponseWriter, r *http.Request) {
	// Parse IDs from URL
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client ID format")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid client object ID format")
		return
	}

	// Restore client object
	response, err := h.clientObjectService.Restore(r.Context(), clientID, id)
	if err != nil {
		if err.Error() == "client not found" {
			WriteNotFound(w, "Client not found")
			return
		}
		if err.Error() == "client object not found" {
			WriteNotFound(w, "Client object not found")
			return
		}
		if err.Error()[:len("client object with name")] == "client object with name" {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to restore client object")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}
