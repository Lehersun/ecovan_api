package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateClientObjectRequest represents the request to create a client object
type CreateClientObjectRequest struct {
	Name    string   `json:"name" validate:"required,min=1,max=255"`
	Address string   `json:"address" validate:"required,min=1,max=500"`
	GeoLat  *float64 `json:"geoLat,omitempty" validate:"omitempty,min=-90,max=90"`
	GeoLng  *float64 `json:"geoLng,omitempty" validate:"omitempty,min=-180,max=180"`
	Notes   *string  `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateClientObjectRequest represents the request to update a client object
type UpdateClientObjectRequest struct {
	Name    string   `json:"name" validate:"required,min=1,max=255"`
	Address string   `json:"address" validate:"required,min=1,max=500"`
	GeoLat  *float64 `json:"geoLat,omitempty" validate:"omitempty,min=-90,max=90"`
	GeoLng  *float64 `json:"geoLng,omitempty" validate:"omitempty,min=-180,max=180"`
	Notes   *string  `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// ClientObjectListRequest represents the request to list client objects
type ClientObjectListRequest struct {
	Page           int  `json:"page" validate:"min=1"`
	PageSize       int  `json:"pageSize" validate:"min=1,max=100"`
	IncludeDeleted bool `json:"includeDeleted"`
}

// ClientObjectListResponse represents the response for listing client objects
type ClientObjectListResponse struct {
	Items    []ClientObjectResponse `json:"items"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
	Total    int64                  `json:"total"`
}

// ClientObjectResponse represents the response for a client object
type ClientObjectResponse struct {
	ID        uuid.UUID  `json:"id"`
	ClientID  uuid.UUID  `json:"clientId"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	GeoLat    *float64   `json:"geoLat,omitempty"`
	GeoLng    *float64   `json:"geoLng,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// DeleteConflicts provides detailed information about what prevents deletion
type DeleteConflicts struct {
	HasActiveOrders    bool        `json:"hasActiveOrders"`
	HasActiveEquipment bool        `json:"hasActiveEquipment"`
	ActiveOrderIDs     []uuid.UUID `json:"activeOrderIds,omitempty"`
	ActiveEquipmentIDs []uuid.UUID `json:"activeEquipmentIds,omitempty"`
	Message            string      `json:"message"`
}

// ToResponse converts ClientObject to ClientObjectResponse
func (co *ClientObject) ToResponse() ClientObjectResponse {
	return ClientObjectResponse{
		ID:        co.ID,
		ClientID:  co.ClientID,
		Name:      co.Name,
		Address:   co.Address,
		GeoLat:    co.GeoLat,
		GeoLng:    co.GeoLng,
		Notes:     co.Notes,
		CreatedAt: co.CreatedAt,
		UpdatedAt: co.UpdatedAt,
		DeletedAt: co.DeletedAt,
	}
}

// FromCreateRequest creates a new ClientObject from CreateClientObjectRequest
func FromCreateClientObjectRequest(clientID uuid.UUID, req CreateClientObjectRequest) *ClientObject {
	now := time.Now()
	return &ClientObject{
		ID:        uuid.New(),
		ClientID:  clientID,
		Name:      req.Name,
		Address:   req.Address,
		GeoLat:    req.GeoLat,
		GeoLng:    req.GeoLng,
		Notes:     req.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateFromRequest updates ClientObject from UpdateClientObjectRequest
func (co *ClientObject) UpdateFromRequest(req UpdateClientObjectRequest) {
	co.Name = req.Name
	co.Address = req.Address
	co.GeoLat = req.GeoLat
	co.GeoLng = req.GeoLng
	co.Notes = req.Notes
	co.UpdatedAt = time.Now()
}

// IsDeleted returns true if the client object is soft deleted
func (co *ClientObject) IsDeleted() bool {
	return co.DeletedAt != nil
}

// SoftDelete marks the client object as deleted
func (co *ClientObject) SoftDelete() {
	now := time.Now()
	co.DeletedAt = &now
	co.UpdatedAt = now
}

// Restore removes the deleted flag from the client object
func (co *ClientObject) Restore() {
	co.DeletedAt = nil
	co.UpdatedAt = time.Now()
}
