package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	orderService port.OrderService
	validate     *validator.Validate
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService port.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		validate:     validator.New(),
	}
}

// ListOrders handles GET /api/v1/orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	date := r.URL.Query().Get("date")
	clientIDStr := r.URL.Query().Get("clientId")
	objectIDStr := r.URL.Query().Get("objectId")
	includeDeleted := r.URL.Query().Get("includeDeleted") == QueryParamIncludeDeleted

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// Build request
	req := models.OrderListRequest{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	}

	// Parse optional filters
	if status != "" {
		orderStatus := models.OrderStatus(status)
		if !models.IsValidOrderStatus(status) {
			WriteValidationError(w, "Invalid status value")
			return
		}
		req.Status = &orderStatus
	}

	if priority != "" {
		if !models.IsValidOrderPriority(priority) {
			WriteValidationError(w, "Invalid priority value")
			return
		}
		req.Priority = &priority
	}

	if date != "" {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			WriteBadRequest(w, "Invalid date format. Expected YYYY-MM-DD")
			return
		}
		req.Date = &parsedDate
	}

	if clientIDStr != "" {
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			WriteBadRequest(w, "Invalid client ID format")
			return
		}
		req.ClientID = &clientID
	}

	if objectIDStr != "" {
		objectID, err := uuid.Parse(objectIDStr)
		if err != nil {
			WriteBadRequest(w, "Invalid object ID format")
			return
		}
		req.ObjectID = &objectID
	}

	// Get orders from service
	response, err := h.orderService.List(r.Context(), req)
	if err != nil {
		WriteInternalError(w, "Failed to list orders")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, response)
}

// GetOrder handles GET /api/v1/orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Get order from service
	order, err := h.orderService.GetByID(r.Context(), orderID)
	if err != nil {
		WriteNotFound(w, "Order not found")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, order)
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateOrderRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Get user ID from context (if available)
	var createdBy *uuid.UUID
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	// Create order
	order, err := h.orderService.Create(r.Context(), &req, createdBy)
	if err != nil {
		WriteInternalError(w, "Failed to create order")
		return
	}

	// Return response
	WriteJSON(w, http.StatusCreated, order)
}

// UpdateOrder handles PUT /api/v1/orders/{id}
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Parse request body
	var req models.UpdateOrderRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Update order
	order, err := h.orderService.Update(r.Context(), orderID, req)
	if err != nil {
		WriteInternalError(w, "Failed to update order")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, order)
}

// UpdateOrderStatus handles PUT /api/v1/orders/{id}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Parse request body
	var req models.UpdateOrderStatusRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Update order status
	order, err := h.orderService.UpdateStatus(r.Context(), orderID, req)
	if err != nil {
		// Check for specific error types
		if err.Error() == "invalid status transition" {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to update order status")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, order)
}

// AssignTransport handles PUT /api/v1/orders/{id}/assign-transport
func (h *OrderHandler) AssignTransport(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Parse request body
	var req models.AssignTransportRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		WriteValidationError(w, err.Error())
		return
	}

	// Assign transport
	err = h.orderService.AssignTransport(r.Context(), orderID, req)
	if err != nil {
		WriteInternalError(w, "Failed to assign transport")
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// DeleteOrder handles DELETE /api/v1/orders/{id}
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Delete order
	err = h.orderService.Delete(r.Context(), orderID)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "cannot be deleted") {
			WriteConflict(w, err.Error())
			return
		}
		WriteInternalError(w, "Failed to delete order")
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// RestoreOrder handles POST /api/v1/orders/{id}/restore
func (h *OrderHandler) RestoreOrder(w http.ResponseWriter, r *http.Request) {
	// Parse order ID
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid order ID")
		return
	}

	// Restore order
	order, err := h.orderService.Restore(r.Context(), orderID)
	if err != nil {
		WriteInternalError(w, "Failed to restore order")
		return
	}

	// Return response
	WriteJSON(w, http.StatusOK, order)
}
