package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Client represents a company client
type Client struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	TaxID     *string    `json:"tax_id,omitempty" db:"tax_id"`
	Email     *string    `json:"email,omitempty" db:"email"`
	Phone     *string    `json:"phone,omitempty" db:"phone"`
	Notes     *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ClientObject represents a physical location/address for a client
type ClientObject struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ClientID  uuid.UUID  `json:"client_id" db:"client_id"`
	Name      string     `json:"name" db:"name"`
	Address   string     `json:"address" db:"address"`
	GeoLat    *float64   `json:"geo_lat,omitempty" db:"geo_lat"`
	GeoLng    *float64   `json:"geo_lng,omitempty" db:"geo_lng"`
	Notes     *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Equipment represents waste bins and containers
type Equipment struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Number         *string    `json:"number,omitempty" db:"number"`
	Type           string     `json:"type" db:"type"`
	VolumeL        int        `json:"volume_l" db:"volume_l"`
	Condition      string     `json:"condition" db:"condition"`
	Photo          *string    `json:"photo,omitempty" db:"photo"`
	ClientObjectID *uuid.UUID `json:"client_object_id,omitempty" db:"client_object_id"`
	WarehouseID    *uuid.UUID `json:"warehouse_id,omitempty" db:"warehouse_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Transport represents vehicles
type Transport struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	PlateNo            string     `json:"plate_no" db:"plate_no"`
	Brand              string     `json:"brand" db:"brand"`
	Model              string     `json:"model" db:"model"`
	CapacityL          int        `json:"capacity_l" db:"capacity_l"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id,omitempty" db:"current_driver_id"`
	CurrentEquipmentID *uuid.UUID `json:"current_equipment_id,omitempty" db:"current_equipment_id"`
	Status             string     `json:"status" db:"status"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Driver represents personnel
type Driver struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	FullName     string     `json:"full_name" db:"full_name"`
	Phone        *string    `json:"phone,omitempty" db:"phone"`
	LicenseNo    string     `json:"license_no" db:"license_no"`
	LicenseClass string     `json:"license_class" db:"license_class"`
	Photo        *string    `json:"photo,omitempty" db:"photo"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Order represents waste collection requests
type Order struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	ClientID            uuid.UUID  `json:"clientId" db:"client_id"`
	ObjectID            uuid.UUID  `json:"objectId" db:"object_id"`
	ScheduledDate       time.Time  `json:"scheduledDate" db:"scheduled_date"`
	ScheduledWindowFrom *string    `json:"scheduledWindowFrom,omitempty" db:"scheduled_window_from"`
	ScheduledWindowTo   *string    `json:"scheduledWindowTo,omitempty" db:"scheduled_window_to"`
	Status              string     `json:"status" db:"status"`
	TransportID         *uuid.UUID `json:"transportId,omitempty" db:"transport_id"`
	Notes               *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy           *uuid.UUID `json:"createdBy,omitempty" db:"created_by"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt           *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

// ToResponse converts an Order model to OrderResponse
func (o *Order) ToResponse() OrderResponse {
	return OrderResponse{
		ID:                  o.ID,
		ClientID:            o.ClientID,
		ObjectID:            o.ObjectID,
		ScheduledDate:       o.ScheduledDate,
		ScheduledWindowFrom: o.ScheduledWindowFrom,
		ScheduledWindowTo:   o.ScheduledWindowTo,
		Status:              o.Status,
		TransportID:         o.TransportID,
		Notes:               o.Notes,
		CreatedBy:           o.CreatedBy,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
		DeletedAt:           o.DeletedAt,
	}
}

// FromCreateRequest creates a new Order from CreateOrderRequest
func FromOrderCreateRequest(req CreateOrderRequest) Order {
	return Order{
		ClientID:            req.ClientID,
		ObjectID:            req.ObjectID,
		ScheduledDate:       req.ScheduledDate,
		ScheduledWindowFrom: req.ScheduledWindowFrom,
		ScheduledWindowTo:   req.ScheduledWindowTo,
		Status:              string(OrderStatusDraft), // Default status
		Notes:               req.Notes,
	}
}

// UpdateFromRequest updates an Order from UpdateOrderRequest
func (o *Order) UpdateFromRequest(req UpdateOrderRequest) {
	if req.ClientID != nil {
		o.ClientID = *req.ClientID
	}
	if req.ObjectID != nil {
		o.ObjectID = *req.ObjectID
	}
	if req.ScheduledDate != nil {
		o.ScheduledDate = *req.ScheduledDate
	}
	if req.ScheduledWindowFrom != nil {
		o.ScheduledWindowFrom = req.ScheduledWindowFrom
	}
	if req.ScheduledWindowTo != nil {
		o.ScheduledWindowTo = req.ScheduledWindowTo
	}
	if req.Notes != nil {
		o.Notes = req.Notes
	}
}

// CanTransitionTo checks if the order can transition to the new status
func (o *Order) CanTransitionTo(newStatus OrderStatus) error {
	switch o.Status {
	case string(OrderStatusDraft):
		if newStatus == OrderStatusScheduled || newStatus == OrderStatusCanceled {
			return nil
		}
		return fmt.Errorf("order in DRAFT status can only transition to SCHEDULED or CANCELED")

	case string(OrderStatusScheduled):
		if newStatus == OrderStatusInProgress || newStatus == OrderStatusCanceled {
			return nil
		}
		return fmt.Errorf("order in SCHEDULED status can only transition to IN_PROGRESS or CANCELED")

	case string(OrderStatusInProgress):
		if newStatus == OrderStatusCompleted {
			return nil
		}
		return fmt.Errorf("order in IN_PROGRESS status can only transition to COMPLETED")

	case string(OrderStatusCompleted):
		return fmt.Errorf("order in COMPLETED status cannot transition to any other status")

	case string(OrderStatusCanceled):
		return fmt.Errorf("order in CANCELED status cannot transition to any other status")

	default:
		return fmt.Errorf("invalid current status: %s", o.Status)
	}
}

// CanBeDeleted checks if the order can be soft-deleted
func (o *Order) CanBeDeleted() error {
	if o.Status == string(OrderStatusDraft) || o.Status == string(OrderStatusCanceled) {
		return nil
	}
	return fmt.Errorf("order in %s status cannot be deleted", o.Status)
}

// AssignTransport assigns transport to the order
func (o *Order) AssignTransport(transportID uuid.UUID) {
	o.TransportID = &transportID
}

// UnassignTransport removes transport assignment from the order
func (o *Order) UnassignTransport() {
	o.TransportID = nil
}

// OrderStatus represents the possible states of an order
type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "DRAFT"
	OrderStatusScheduled  OrderStatus = "SCHEDULED"
	OrderStatusInProgress OrderStatus = "IN_PROGRESS"
	OrderStatusCompleted  OrderStatus = "COMPLETED"
	OrderStatusCanceled   OrderStatus = "CANCELED"
)

// ValidOrderStatuses contains all valid order statuses
var ValidOrderStatuses = []OrderStatus{
	OrderStatusDraft,
	OrderStatusScheduled,
	OrderStatusInProgress,
	OrderStatusCompleted,
	OrderStatusCanceled,
}

// IsValidOrderStatus checks if a status is valid
func IsValidOrderStatus(status string) bool {
	for _, validStatus := range ValidOrderStatuses {
		if string(validStatus) == status {
			return true
		}
	}
	return false
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	ClientID            uuid.UUID `json:"clientId" validate:"required"`
	ObjectID            uuid.UUID `json:"objectId" validate:"required"`
	ScheduledDate       time.Time `json:"scheduledDate" validate:"required"`
	ScheduledWindowFrom *string   `json:"scheduledWindowFrom,omitempty"`
	ScheduledWindowTo   *string   `json:"scheduledWindowTo,omitempty"`
	Notes               *string   `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateOrderRequest represents the request to update an existing order
type UpdateOrderRequest struct {
	ClientID            *uuid.UUID `json:"clientId,omitempty" validate:"omitempty"`
	ObjectID            *uuid.UUID `json:"objectId,omitempty" validate:"omitempty"`
	ScheduledDate       *time.Time `json:"scheduledDate,omitempty" validate:"omitempty"`
	ScheduledWindowFrom *string    `json:"scheduledWindowFrom,omitempty"`
	ScheduledWindowTo   *string    `json:"scheduledWindowTo,omitempty"`
	Notes               *string    `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=DRAFT SCHEDULED IN_PROGRESS COMPLETED CANCELED"`
}

// AssignTransportRequest represents the request to assign transport to an order
type AssignTransportRequest struct {
	TransportID uuid.UUID `json:"transportId" validate:"required"`
}

// OrderListRequest represents the request to list orders with filtering and pagination
type OrderListRequest struct {
	Page           int          `json:"page" validate:"min=1"`
	PageSize       int          `json:"pageSize" validate:"min=1,max=100"`
	Status         *OrderStatus `json:"status,omitempty"`
	Date           *time.Time   `json:"date,omitempty"`
	ClientID       *uuid.UUID   `json:"clientId,omitempty"`
	ObjectID       *uuid.UUID   `json:"objectId,omitempty"`
	IncludeDeleted bool         `json:"includeDeleted"`
}

// OrderListResponse represents the paginated response for listing orders
type OrderListResponse struct {
	Items    []OrderResponse `json:"items"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Total    int64           `json:"total"`
}

// OrderResponse represents a single order response
type OrderResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ClientID            uuid.UUID  `json:"clientId"`
	ObjectID            uuid.UUID  `json:"objectId"`
	ScheduledDate       time.Time  `json:"scheduledDate"`
	ScheduledWindowFrom *string    `json:"scheduledWindowFrom,omitempty"`
	ScheduledWindowTo   *string    `json:"scheduledWindowTo,omitempty"`
	Status              string     `json:"status"`
	TransportID         *uuid.UUID `json:"transportId,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
	CreatedBy           *uuid.UUID `json:"createdBy,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	DeletedAt           *time.Time `json:"deletedAt,omitempty"`
}
