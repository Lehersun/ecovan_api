package models

import (
	"encoding/json"
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
	Brand     string          `json:"brand" validate:"required,min=1,max=50"`
	Model     string          `json:"model" validate:"required,min=1,max=50"`
	CapacityL int             `json:"capacityL" validate:"required,gt=0"`
	Status    TransportStatus `json:"status,omitempty" validate:"omitempty,oneof=IN_WORK REPAIR"`
	DriverID  *uuid.UUID      `json:"driverId,omitempty" validate:"omitempty"`
}

// UpdateTransportRequest represents the request to update an existing transport
type UpdateTransportRequest struct {
	PlateNo   *string          `json:"plateNo,omitempty" validate:"omitempty,min=1,max=20"`
	Brand     *string          `json:"brand,omitempty" validate:"omitempty,min=1,max=50"`
	Model     *string          `json:"model,omitempty" validate:"omitempty,min=1,max=50"`
	CapacityL *int             `json:"capacityL,omitempty" validate:"omitempty,gt=0"`
	Status    *TransportStatus `json:"status,omitempty" validate:"omitempty,oneof=IN_WORK REPAIR"`
	DriverID  *uuid.UUID       `json:"driverId,omitempty" validate:"omitempty"`

	// Internal fields to track explicit null values
	DriverIDExplicitlySet  bool
	DriverIDExplicitlyNull bool
}

// UnmarshalJSON custom unmarshaler to handle explicit null values
func (req *UpdateTransportRequest) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to unmarshal into
	type tempUpdateTransportRequest struct {
		PlateNo   *string          `json:"plateNo,omitempty"`
		Brand     *string          `json:"brand,omitempty"`
		Model     *string          `json:"model,omitempty"`
		CapacityL *int             `json:"capacityL,omitempty"`
		Status    *TransportStatus `json:"status,omitempty"`
		DriverID  *uuid.UUID       `json:"driverId,omitempty"`
	}

	var temp tempUpdateTransportRequest
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy the fields
	req.PlateNo = temp.PlateNo
	req.Brand = temp.Brand
	req.Model = temp.Model
	req.CapacityL = temp.CapacityL
	req.Status = temp.Status
	req.DriverID = temp.DriverID

	// Check if driverId was explicitly set in the JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, exists := raw["driverId"]; exists {
			req.DriverIDExplicitlySet = true
			if raw["driverId"] == nil {
				req.DriverIDExplicitlyNull = true
			}
		}
	}

	return nil
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
	Brand              string     `json:"brand"`
	Model              string     `json:"model"`
	CapacityL          int        `json:"capacityL"`
	CurrentDriverID    *uuid.UUID `json:"currentDriverId"`
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
		Brand:              t.Brand,
		Model:              t.Model,
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
func FromTransportCreateRequest(req *CreateTransportRequest) Transport {
	now := time.Now()
	status := req.Status
	if status == "" {
		status = TransportStatusInWork
	}

	return Transport{
		PlateNo:         req.PlateNo,
		Brand:           req.Brand,
		Model:           req.Model,
		CapacityL:       req.CapacityL,
		Status:          string(status),
		CurrentDriverID: req.DriverID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// UpdateFromRequest updates a Transport from UpdateTransportRequest
func (t *Transport) UpdateFromRequest(req UpdateTransportRequest) {
	if req.PlateNo != nil {
		t.PlateNo = *req.PlateNo
	}
	if req.Brand != nil {
		t.Brand = *req.Brand
	}
	if req.Model != nil {
		t.Model = *req.Model
	}
	if req.CapacityL != nil {
		t.CapacityL = *req.CapacityL
	}
	if req.Status != nil {
		t.Status = string(*req.Status)
	}
	// Handle driver assignment/unassignment
	if req.DriverIDExplicitlySet {
		if req.DriverIDExplicitlyNull {
			// Explicitly set to null (unassign)
			t.CurrentDriverID = nil
		} else if req.DriverID != nil {
			// Assign specific driver
			t.CurrentDriverID = req.DriverID
		}
	}
	t.UpdatedAt = time.Now()
}
