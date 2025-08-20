package http

import (
	"eco-van-api/internal/models"
	"eco-van-api/internal/port"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type DriverHandler struct {
	driverService port.DriverService
	validate      *validator.Validate
}

// NewDriverHandler creates a new driver handler
func NewDriverHandler(driverService port.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
		validate:      validator.New(),
	}
}

// ListDrivers handles GET /api/v1/drivers
func (h *DriverHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
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

	licenseClass := r.URL.Query().Get("licenseClass")
	var licenseClassPtr *string
	if licenseClass != "" {
		licenseClassPtr = &licenseClass
	}

	q := r.URL.Query().Get("q")
	var qPtr *string
	if q != "" {
		qPtr = &q
	}

	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Build request
	req := models.DriverListRequest{
		Page:           page,
		PageSize:       pageSize,
		LicenseClass:   licenseClassPtr,
		Q:              qPtr,
		IncludeDeleted: includeDeleted,
	}

	// Call service
	drivers, err := h.driverService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list drivers")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, drivers)
}

// GetDriver handles GET /api/v1/drivers/{id}
func (h *DriverHandler) GetDriver(w http.ResponseWriter, r *http.Request) {
	// Parse driver ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid driver ID")
		return
	}

	// Call service
	driver, err := h.driverService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Driver not found")
			return
		}
		WriteInternalError(w, "Failed to get driver")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, driver)
}

// CreateDriver handles POST /api/v1/drivers
func (h *DriverHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateDriverRequest
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
	driver, err := h.driverService.Create(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Driver with this license number already exists")
			return
		}
		WriteInternalError(w, "Failed to create driver")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, driver)
}

// UpdateDriver handles PUT /api/v1/drivers/{id}
func (h *DriverHandler) UpdateDriver(w http.ResponseWriter, r *http.Request) {
	// Parse driver ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid driver ID")
		return
	}

	// Parse request body
	var req models.UpdateDriverRequest
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
	driver, err := h.driverService.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Driver not found")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			WriteConflict(w, "Driver with this license number already exists")
			return
		}
		WriteInternalError(w, "Failed to update driver")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, driver)
}

// DeleteDriver handles DELETE /api/v1/drivers/{id}
func (h *DriverHandler) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	// Parse driver ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid driver ID")
		return
	}

	// Call service
	err = h.driverService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "while assigned to transport") {
			WriteConflict(w, "Cannot delete driver while assigned to transport")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Driver not found")
			return
		}
		WriteInternalError(w, "Failed to delete driver")
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// RestoreDriver handles POST /api/v1/drivers/{id}/restore
func (h *DriverHandler) RestoreDriver(w http.ResponseWriter, r *http.Request) {
	// Parse driver ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid driver ID")
		return
	}

	// Call service
	driver, err := h.driverService.Restore(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFound(w, "Driver not found")
			return
		}
		if strings.Contains(err.Error(), "not soft-deleted") {
			WriteBadRequest(w, "Driver is not soft-deleted")
			return
		}
		WriteInternalError(w, "Failed to restore driver")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, driver)
}

// ListAvailableDrivers handles GET /api/v1/drivers/available
func (h *DriverHandler) ListAvailableDrivers(w http.ResponseWriter, r *http.Request) {
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

	licenseClass := r.URL.Query().Get("licenseClass")
	var licenseClassPtr *string
	if licenseClass != "" {
		licenseClassPtr = &licenseClass
	}

	q := r.URL.Query().Get("q")
	var qPtr *string
	if q != "" {
		qPtr = &q
	}

	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	// Build request
	req := models.DriverListRequest{
		Page:           page,
		PageSize:       pageSize,
		LicenseClass:   licenseClassPtr,
		Q:              qPtr,
		IncludeDeleted: includeDeleted,
	}

	// Call service
	drivers, err := h.driverService.ListAvailable(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list available drivers")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, drivers)
}
