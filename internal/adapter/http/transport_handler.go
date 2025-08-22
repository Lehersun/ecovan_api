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

// transportHandler handles HTTP requests for transport operations
type transportHandler struct {
	transportService port.TransportService
	validate         *validator.Validate
}

// NewTransportHandler creates a new transport handler
func NewTransportHandler(transportService port.TransportService) *transportHandler {
	return &transportHandler{
		transportService: transportService,
		validate:         validator.New(),
	}
}

// ListItems handles GET /v1/transport with filtering and pagination
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *transportHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	status := r.URL.Query().Get("status")
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Build request
	req := models.TransportListRequest{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	}

	if status != "" {
		req.Status = &status
	}

	// Get transport list
	response, err := h.transportService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// GetItem handles GET /v1/transport/{id}
func (h *transportHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	// Get transport by ID
	transport, err := h.transportService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Transport not found")
			return
		}
		WriteInternalError(w, "Failed to get transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, transport)
}

// CreateItem handles POST /v1/transport
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *transportHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransportRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Create transport
	transport, err := h.transportService.Create(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Transport with this plate number already exists")
			return
		}
		WriteInternalError(w, "Failed to create transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, transport)
}

// UpdateItem handles PUT /v1/transport/{id}
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *transportHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	var req models.UpdateTransportRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Update transport
	transport, err := h.transportService.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Transport not found")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Transport with this plate number already exists")
			return
		}
		WriteInternalError(w, "Failed to update transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, transport)
}

// DeleteItem handles DELETE /v1/transport/{id}
func (h *transportHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	// Delete transport
	err = h.transportService.Delete(r.Context(), id)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "driver is currently assigned") ||
			strings.Contains(err.Error(), "equipment is currently assigned") ||
			strings.Contains(err.Error(), "has active orders") {
			WriteConflict(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Transport not found")
			return
		}
		WriteInternalError(w, "Failed to delete transport")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RestoreItem handles POST /v1/transport/{id}/restore
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *transportHandler) RestoreItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	// Restore transport
	transport, err := h.transportService.Restore(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Transport not found")
			return
		}
		WriteInternalError(w, "Failed to restore transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, transport)
}

// GetAvailable handles GET /v1/transport/available
//
//nolint:dupl // Similar pattern but with different service calls and filters
func (h *transportHandler) GetAvailable(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Build request
	req := models.TransportListRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// Get available transport
	response, err := h.transportService.GetAvailable(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to get available transport")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// AssignDriver handles PUT /v1/transport/{id}/assign-driver
func (h *transportHandler) AssignDriver(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	var req models.AssignDriverRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Assign driver
	err = h.transportService.AssignDriver(r.Context(), idStr, req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "already assigned to another transport") ||
			strings.Contains(err.Error(), "not found") {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to assign driver")
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}

// AssignEquipment handles PUT /v1/transport/{id}/assign-equipment
func (h *transportHandler) AssignEquipment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	var req models.AssignEquipmentRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Assign equipment
	err = h.transportService.AssignEquipment(r.Context(), idStr, req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "not available for assignment") ||
			strings.Contains(err.Error(), "already assigned to another transport") ||
			strings.Contains(err.Error(), "not found") {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to assign equipment")
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}

// UnassignDriver handles DELETE /v1/transport/{id}/drivers
func (h *transportHandler) UnassignDriver(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid transport ID")
		return
	}

	// Unassign driver
	err = h.transportService.UnassignDriver(r.Context(), idStr)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "no driver assigned") {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to unassign driver")
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}
