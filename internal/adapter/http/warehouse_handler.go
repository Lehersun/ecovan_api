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

// warehouseHandler handles HTTP requests for warehouse operations
type warehouseHandler struct {
	warehouseService port.WarehouseService
	validate         *validator.Validate
}

// NewWarehouseHandler creates a new warehouse handler
func NewWarehouseHandler(warehouseService port.WarehouseService) *warehouseHandler {
	return &warehouseHandler{
		warehouseService: warehouseService,
		validate:         validator.New(),
	}
}

// RegisterRoutes registers the warehouse routes with the router
func (h *warehouseHandler) RegisterRoutes(router chi.Router) {
	router.Get("/warehouses", h.ListWarehouses)
	router.Post("/warehouses", h.CreateWarehouse)
	router.Get("/warehouses/{id}", h.GetWarehouse)
	router.Put("/warehouses/{id}", h.UpdateWarehouse)
	router.Delete("/warehouses/{id}", h.DeleteWarehouse)
	router.Post("/warehouses/{id}/restore", h.RestoreWarehouse)
}

// ListWarehouses handles GET /v1/warehouses with pagination and filtering
func (h *warehouseHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// Create request
	req := models.WarehouseListRequest{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	}

	// Get warehouses from service
	response, err := h.warehouseService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list warehouses")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// CreateWarehouse handles POST /v1/warehouses
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *warehouseHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWarehouseRequest

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

	// Create warehouse via service
	response, err := h.warehouseService.Create(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Warehouse creation failed")
			return
		}
		WriteInternalError(w, "Failed to create warehouse")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, response)
}

// GetWarehouse handles GET /v1/warehouses/{id}
func (h *warehouseHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	// Parse warehouse ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid warehouse ID")
		return
	}

	// Get warehouse from service
	warehouse, err := h.warehouseService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Warehouse not found")
			return
		}
		WriteInternalError(w, "Failed to get warehouse")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, warehouse)
}

// UpdateWarehouse handles PUT /v1/warehouses/{id}
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *warehouseHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	// Parse warehouse ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid warehouse ID")
		return
	}

	var req models.UpdateWarehouseRequest

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

	// Update warehouse via service
	response, err := h.warehouseService.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Warehouse not found")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Warehouse update failed")
			return
		}
		WriteInternalError(w, "Failed to update warehouse")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// DeleteWarehouse handles DELETE /v1/warehouses/{id}
func (h *warehouseHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	// Parse warehouse ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid warehouse ID")
		return
	}

	// Delete warehouse via service
	err = h.warehouseService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Warehouse not found")
			return
		}
		if strings.Contains(err.Error(), "equipment is still present") {
			WriteCustomProblem(w,
				"/errors/conflict",
				"Conflict",
				http.StatusConflict,
				"Cannot delete warehouse: equipment is still present",
				"/warehouses/"+idStr)
			return
		}
		WriteInternalError(w, "Failed to delete warehouse")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RestoreWarehouse handles POST /v1/warehouses/{id}/restore
//
//nolint:dupl // Similar pattern across handlers but with different business logic
func (h *warehouseHandler) RestoreWarehouse(w http.ResponseWriter, r *http.Request) {
	// Parse warehouse ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid warehouse ID")
		return
	}

	// Restore warehouse via service
	response, err := h.warehouseService.Restore(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Warehouse not found")
			return
		}
		if strings.Contains(err.Error(), "not deleted") {
			WriteBadRequest(w, "Warehouse is not deleted")
			return
		}
		if strings.Contains(err.Error(), "conflicts with existing warehouse") {
			WriteConflict(w, "Cannot restore warehouse: name conflicts with existing warehouse")
			return
		}
		WriteInternalError(w, "Failed to restore warehouse")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}
