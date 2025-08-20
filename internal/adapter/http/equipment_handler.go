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

// EquipmentHandler handles HTTP requests for equipment operations
type EquipmentHandler struct {
	equipmentService port.EquipmentService
	validate         *validator.Validate
}

// NewEquipmentHandler creates a new equipment handler
func NewEquipmentHandler(equipmentService port.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{
		equipmentService: equipmentService,
		validate:         validator.New(),
	}
}

// ListEquipment handles GET /api/v1/equipment
func (h *EquipmentHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize <= 0 {
		pageSize = 20
	}

	includeDeleted, _ := strconv.ParseBool(r.URL.Query().Get("includeDeleted"))

	// Parse type filter
	var equipmentType *models.EquipmentType
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := models.EquipmentType(typeStr)
		equipmentType = &t
	}

	// Parse client object filter
	var clientObjectID *uuid.UUID
	if clientObjectIDStr := r.URL.Query().Get("clientObjectId"); clientObjectIDStr != "" {
		if id, err := uuid.Parse(clientObjectIDStr); err == nil {
			clientObjectID = &id
		}
	}

	// Parse warehouse filter
	var warehouseID *uuid.UUID
	if warehouseIDStr := r.URL.Query().Get("warehouseId"); warehouseIDStr != "" {
		if id, err := uuid.Parse(warehouseIDStr); err == nil {
			warehouseID = &id
		}
	}

	// Build request
	req := models.EquipmentListRequest{
		Page:           page,
		PageSize:       pageSize,
		Type:           equipmentType,
		ClientObjectID: clientObjectID,
		WarehouseID:    warehouseID,
		IncludeDeleted: includeDeleted,
	}

	// Call service
	response, err := h.equipmentService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list equipment")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// GetEquipment handles GET /api/v1/equipment/{id}
func (h *EquipmentHandler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid equipment ID")
		return
	}

	// Call service
	equipment, err := h.equipmentService.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == ErrEquipmentNotFound {
			WriteNotFound(w, "Equipment not found")
			return
		}
		WriteInternalError(w, "Failed to get equipment")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, equipment)
}

// CreateEquipment handles POST /api/v1/equipment
func (h *EquipmentHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateEquipmentRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Call service
	equipment, err := h.equipmentService.Create(r.Context(), req)
	if err != nil {
		if err.Error() == "validation failed: equipment cannot be placed at both client object and warehouse simultaneously" {
			WriteProblemWithType(w, http.StatusUnprocessableEntity,
				"/errors/unprocessable-entity",
				"Cannot place equipment at both client object and warehouse simultaneously")
			return
		}
		if err.Error() == "equipment with number '"+*req.Number+"' already exists" {
			WriteConflict(w, "Equipment with this number already exists")
			return
		}
		WriteInternalError(w, "Failed to create equipment")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, equipment)
}

// UpdateEquipment handles PUT /api/v1/equipment/{id}
func (h *EquipmentHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid equipment ID")
		return
	}

	// Parse request body
	var req models.UpdateEquipmentRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, "Validation failed")
		return
	}

	// Call service
	equipment, err := h.equipmentService.Update(r.Context(), id, req)
	if err != nil {
		if err.Error() == "validation failed: equipment cannot be placed at both client object and warehouse simultaneously" {
			WriteProblemWithType(w, http.StatusUnprocessableEntity,
				"/errors/unprocessable-entity",
				"Cannot place equipment at both client object and warehouse simultaneously")
			return
		}
		if err.Error() == ErrEquipmentNotFound {
			WriteNotFound(w, "Equipment not found")
			return
		}
		if err.Error() == "cannot change equipment placement while attached to transport" {
			WriteProblemWithType(w, http.StatusConflict,
				"/errors/conflict",
				"Cannot change equipment placement while attached to transport")
			return
		}
		if err.Error() == "equipment with number '"+*req.Number+"' already exists" {
			WriteConflict(w, "Equipment with this number already exists")
			return
		}
		WriteInternalError(w, "Failed to update equipment")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, equipment)
}

// DeleteEquipment handles DELETE /api/v1/equipment/{id}
func (h *EquipmentHandler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid equipment ID")
		return
	}

	// Call service
	err = h.equipmentService.Delete(r.Context(), id)
	if err != nil {
		if err.Error() == ErrEquipmentNotFound {
			WriteNotFound(w, "Equipment not found")
			return
		}
		if err.Error() == "cannot delete equipment while attached to transport" {
			WriteProblemWithType(w, http.StatusConflict,
				"/errors/conflict",
				"Cannot delete equipment while attached to transport")
			return
		}
		WriteInternalError(w, "Failed to delete equipment")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RestoreEquipment handles POST /api/v1/equipment/{id}/restore
func (h *EquipmentHandler) RestoreEquipment(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid equipment ID")
		return
	}

	// Call service
	equipment, err := h.equipmentService.Restore(r.Context(), id)
	if err != nil {
		if err.Error() == ErrEquipmentNotFound {
			WriteNotFound(w, "Equipment not found")
			return
		}
		if err.Error() == "equipment is not deleted" {
			WriteBadRequest(w, "Equipment is not deleted")
			return
		}
		if err.Error() == "cannot restore equipment: number '"+*equipment.Number+"' conflicts with existing equipment" {
			WriteConflict(w, "Cannot restore equipment: number conflicts with existing equipment")
			return
		}
		WriteInternalError(w, "Failed to restore equipment")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, equipment)
}
