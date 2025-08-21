package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EquipmentType represents the type of equipment
type EquipmentType string

const (
	EquipmentTypeBin       EquipmentType = "BIN"
	EquipmentTypeContainer EquipmentType = "CONTAINER"
)

// EquipmentCondition represents the condition of equipment
type EquipmentCondition string

const (
	EquipmentConditionGood         EquipmentCondition = "GOOD"
	EquipmentConditionDamaged      EquipmentCondition = "DAMAGED"
	EquipmentConditionOutOfService EquipmentCondition = "OUT_OF_SERVICE"
)

// Error constants
var ErrInvalidPlacement = fmt.Errorf("equipment must be assigned to exactly one of: transport, client object, or warehouse")

// CreateEquipmentRequest represents the request to create new equipment
type CreateEquipmentRequest struct {
	Number         *string            `json:"number,omitempty" validate:"omitempty,max=100"`
	Type           EquipmentType      `json:"type" validate:"required,oneof=BIN CONTAINER"`
	VolumeL        int                `json:"volumeL" validate:"required,gt=0"`
	Condition      EquipmentCondition `json:"condition" validate:"required,oneof=GOOD DAMAGED OUT_OF_SERVICE"`
	Photo          *string            `json:"photo,omitempty" validate:"omitempty,max=500"`
	ClientObjectID *uuid.UUID         `json:"clientObjectId"`
	WarehouseID    *uuid.UUID         `json:"warehouseId"`
	TransportID    *uuid.UUID         `json:"transportId"`
}

// UpdateEquipmentRequest represents the request to update existing equipment
type UpdateEquipmentRequest struct {
	Number         *string            `json:"number,omitempty" validate:"omitempty,max=100"`
	Type           EquipmentType      `json:"type" validate:"required,oneof=BIN CONTAINER"`
	VolumeL        int                `json:"volumeL" validate:"required,gt=0"`
	Condition      EquipmentCondition `json:"condition" validate:"required,oneof=GOOD DAMAGED OUT_OF_SERVICE"`
	Photo          *string            `json:"photo,omitempty" validate:"omitempty,max=500"`
	ClientObjectID *uuid.UUID         `json:"clientObjectId"`
	WarehouseID    *uuid.UUID         `json:"warehouseId"`
	TransportID    *uuid.UUID         `json:"transportId"`
}

// EquipmentListRequest represents the request to list equipment with filtering and pagination
type EquipmentListRequest struct {
	Page           int            `json:"page" validate:"min=1"`
	PageSize       int            `json:"pageSize" validate:"min=1,max=100"`
	Type           *EquipmentType `json:"type,omitempty"`
	ClientObjectID *uuid.UUID     `json:"clientObjectId,omitempty"`
	WarehouseID    *uuid.UUID     `json:"warehouseId,omitempty"`
	TransportID    *uuid.UUID     `json:"transportId,omitempty"`
	IncludeDeleted bool           `json:"includeDeleted"`
}

// EquipmentListResponse represents the paginated response for listing equipment
type EquipmentListResponse struct {
	Items    []EquipmentResponse `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int64               `json:"total"`
}

// EquipmentResponse represents a single equipment response
type EquipmentResponse struct {
	ID             uuid.UUID          `json:"id"`
	Number         *string            `json:"number,omitempty"`
	Type           EquipmentType      `json:"type"`
	VolumeL        int                `json:"volumeL"`
	Condition      EquipmentCondition `json:"condition"`
	Photo          *string            `json:"photo,omitempty"`
	ClientObjectID *uuid.UUID         `json:"clientObjectId"`
	WarehouseID    *uuid.UUID         `json:"warehouseId"`
	TransportID    *uuid.UUID         `json:"transportId"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	DeletedAt      *time.Time         `json:"deletedAt,omitempty"`
}

// ToResponse converts an Equipment model to EquipmentResponse
func (e *Equipment) ToResponse() EquipmentResponse {
	return EquipmentResponse{
		ID:             e.ID,
		Number:         e.Number,
		Type:           EquipmentType(e.Type),
		VolumeL:        e.VolumeL,
		Condition:      EquipmentCondition(e.Condition),
		Photo:          e.Photo,
		ClientObjectID: e.ClientObjectID,
		WarehouseID:    e.WarehouseID,
		TransportID:    e.TransportID,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		DeletedAt:      e.DeletedAt,
	}
}

// FromCreateRequest creates a new Equipment from CreateEquipmentRequest
//
//nolint:gocritic // hugeParam: Interface requires request by value
func FromEquipmentCreateRequest(req CreateEquipmentRequest) Equipment {
	return Equipment{
		Type:           string(req.Type),
		VolumeL:        req.VolumeL,
		Condition:      string(req.Condition),
		Number:         req.Number,
		Photo:          req.Photo,
		ClientObjectID: req.ClientObjectID,
		WarehouseID:    req.WarehouseID,
		TransportID:    req.TransportID,
	}
}

// UpdateFromRequest updates an Equipment from UpdateEquipmentRequest
//
//nolint:gocritic // hugeParam: Interface requires request by value
func (e *Equipment) UpdateFromRequest(req UpdateEquipmentRequest) {
	e.Type = string(req.Type)
	e.VolumeL = req.VolumeL
	e.Condition = string(req.Condition)
	e.Number = req.Number
	e.Photo = req.Photo
	e.ClientObjectID = req.ClientObjectID
	e.WarehouseID = req.WarehouseID
	e.TransportID = req.TransportID
}

// ValidatePlacement validates that exactly one of TransportID, ClientObjectID, or WarehouseID is set
func (req *CreateEquipmentRequest) ValidatePlacement() error {
	count := 0
	if req.TransportID != nil {
		count++
	}
	if req.ClientObjectID != nil {
		count++
	}
	if req.WarehouseID != nil {
		count++
	}

	if count != 1 {
		return fmt.Errorf("equipment must be assigned to exactly one of: transport, client object, or warehouse")
	}
	return nil
}

// ValidatePlacement validates that exactly one of TransportID, ClientObjectID, or WarehouseID is set
func (req *UpdateEquipmentRequest) ValidatePlacement() error {
	count := 0
	if req.TransportID != nil {
		count++
	}
	if req.ClientObjectID != nil {
		count++
	}
	if req.WarehouseID != nil {
		count++
	}

	// For updates, if no placement is specified, it's valid (will keep existing placement)
	// If placement is specified, exactly one must be set
	if count > 0 && count != 1 {
		return fmt.Errorf("equipment must be assigned to exactly one of: transport, client object, or warehouse")
	}
	return nil
}
