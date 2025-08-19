package models

import (
	"time"

	"github.com/google/uuid"
)

// Warehouse represents a storage facility for equipment
type Warehouse struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Address   *string    `json:"address,omitempty" db:"address"`
	Notes     *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateWarehouseRequest represents the request to create a new warehouse
type CreateWarehouseRequest struct {
	Name    string  `json:"name" validate:"required,min=1,max=255"`
	Address *string `json:"address,omitempty" validate:"omitempty,max=500"`
	Notes   *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateWarehouseRequest represents the request to update an existing warehouse
type UpdateWarehouseRequest struct {
	Name    string  `json:"name" validate:"required,min=1,max=255"`
	Address *string `json:"address,omitempty" validate:"omitempty,max=500"`
	Notes   *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// WarehouseListRequest represents the request to list warehouses with filtering and pagination
type WarehouseListRequest struct {
	Page           int  `json:"page" validate:"min=1"`
	PageSize       int  `json:"pageSize" validate:"min=1,max=100"`
	IncludeDeleted bool `json:"includeDeleted"`
}

// WarehouseListResponse represents the paginated response for listing warehouses
type WarehouseListResponse struct {
	Items    []Warehouse `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int64       `json:"total"`
}

// WarehouseResponse represents a single warehouse response
type WarehouseResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Address   *string    `json:"address,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToResponse converts a Warehouse model to WarehouseResponse
func (w *Warehouse) ToResponse() WarehouseResponse {
	return WarehouseResponse{
		ID:        w.ID,
		Name:      w.Name,
		Address:   w.Address,
		Notes:     w.Notes,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		DeletedAt: w.DeletedAt,
	}
}

// FromCreateRequest creates a new Warehouse from CreateWarehouseRequest
func FromWarehouseCreateRequest(req CreateWarehouseRequest) Warehouse {
	return Warehouse{
		Name:    req.Name,
		Address: req.Address,
		Notes:   req.Notes,
	}
}

// UpdateFromRequest updates a Warehouse from UpdateWarehouseRequest
func (w *Warehouse) UpdateFromRequest(req UpdateWarehouseRequest) {
	w.Name = req.Name
	w.Address = req.Address
	w.Notes = req.Notes
}
