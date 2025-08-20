package models

import (
	"time"

	"github.com/google/uuid"
)

// TransportStatus represents the status of transport
type TransportStatus string

const (
	TransportStatusInWork TransportStatus = "IN_WORK"
	TransportStatusRepair TransportStatus = "REPAIR"
)

// CreateTransportRequest represents the request to create a new transport
type CreateTransportRequest struct {
	PlateNo   string          `json:"plateNo" validate:"required,min=1,max=20"`
	CapacityL int             `json:"capacityL" validate:"required,gt=0"`
	Status    TransportStatus `json:"status,omitempty" validate:"omitempty,oneof=IN_WORK REPAIR"`
}

// UpdateTransportRequest represents the request to update an existing transport
type UpdateTransportRequest struct {
	PlateNo   *string          `json:"plateNo,omitempty" validate:"omitempty,min=1,max=20"`
	CapacityL *int             `json:"capacityL,omitempty" validate:"omitempty,gt=0"`
	Status    *TransportStatus `json:"status,omitempty" validate:"omitempty,oneof=IN_WORK REPAIR"`
}

// TransportListRequest represents the request to list transport with filtering and pagination
type TransportListRequest struct {
	Page           int     `json:"page" validate:"min=1"`
	PageSize       int     `json:"pageSize" validate:"min=1,max=100"`
	Status         *string `json:"status,omitempty"`
	IncludeDeleted bool    `json:"includeDeleted"`
}

// TransportListResponse represents the paginated response for listing transport
type TransportListResponse struct {
	Items    []TransportResponse `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int64               `json:"total"`
}

// TransportResponse represents a single transport response
type TransportResponse struct {
	ID                 uuid.UUID  `json:"id"`
	PlateNo            string     `json:"plateNo"`
	CapacityL          int        `json:"capacityL"`
	CurrentDriverID    *uuid.UUID `json:"currentDriverId,omitempty"`
	CurrentEquipmentID *uuid.UUID `json:"currentEquipmentId,omitempty"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt,omitempty"`
}

// AssignDriverRequest represents the request to assign a driver to transport
type AssignDriverRequest struct {
	DriverID uuid.UUID `json:"driverId" validate:"required"`
}

// AssignEquipmentRequest represents the request to assign equipment to transport
type AssignEquipmentRequest struct {
	EquipmentID uuid.UUID `json:"equipmentId" validate:"required"`
}

// ToResponse converts a Transport model to TransportResponse
func (t *Transport) ToResponse() TransportResponse {
	return TransportResponse{
		ID:                 t.ID,
		PlateNo:            t.PlateNo,
		CapacityL:          t.CapacityL,
		CurrentDriverID:    t.CurrentDriverID,
		CurrentEquipmentID: t.CurrentEquipmentID,
		Status:             t.Status,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
		DeletedAt:          t.DeletedAt,
	}
}

// FromCreateRequest creates a new Transport from CreateTransportRequest
func FromTransportCreateRequest(req CreateTransportRequest) Transport {
	now := time.Now()
	status := req.Status
	if status == "" {
		status = TransportStatusInWork
	}

	return Transport{
		PlateNo:   req.PlateNo,
		CapacityL: req.CapacityL,
		Status:    string(status),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateFromRequest updates a Transport from UpdateTransportRequest
func (t *Transport) UpdateFromRequest(req UpdateTransportRequest) {
	if req.PlateNo != nil {
		t.PlateNo = *req.PlateNo
	}
	if req.CapacityL != nil {
		t.CapacityL = *req.CapacityL
	}
	if req.Status != nil {
		t.Status = string(*req.Status)
	}
	t.UpdatedAt = time.Now()
}
