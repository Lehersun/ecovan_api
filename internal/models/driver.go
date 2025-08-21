package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverLicenseClass represents the class of driver's license
type DriverLicenseClass string

const (
	DriverLicenseClassA DriverLicenseClass = "A" // Motorcycle
	DriverLicenseClassB DriverLicenseClass = "B" // Car
	DriverLicenseClassC DriverLicenseClass = "C" // Truck
	DriverLicenseClassD DriverLicenseClass = "D" // Bus
	DriverLicenseClassE DriverLicenseClass = "E" // Trailer
)

// CreateDriverRequest represents the request to create a new driver
type CreateDriverRequest struct {
	FullName     string             `json:"fullName" validate:"required,min=2,max=100"`
	Phone        *string            `json:"phone,omitempty" validate:"omitempty,max=20"`
	LicenseNo    string             `json:"licenseNo" validate:"required,min=5,max=20"`
	LicenseClass DriverLicenseClass `json:"licenseClass" validate:"required,oneof=A B C D E"`
	Photo        *string            `json:"photo,omitempty" validate:"omitempty,max=500"`
}

// UpdateDriverRequest represents the request to update an existing driver
type UpdateDriverRequest struct {
	FullName     *string             `json:"fullName,omitempty" validate:"omitempty,min=2,max=100"`
	Phone        *string             `json:"phone,omitempty" validate:"omitempty,max=20"`
	LicenseNo    *string             `json:"licenseNo,omitempty" validate:"omitempty,min=5,max=20"`
	LicenseClass *DriverLicenseClass `json:"licenseClass,omitempty" validate:"omitempty,oneof=A B C D E"`
	Photo        *string             `json:"photo,omitempty" validate:"omitempty,max=500"`
}

// DriverListRequest represents the request to list drivers with filtering and pagination
type DriverListRequest struct {
	Page           int     `json:"page" validate:"required,min=1"`
	PageSize       int     `json:"pageSize" validate:"required,min=1,max=100"`
	LicenseClass   *string `json:"licenseClass,omitempty"`
	Q              *string `json:"q,omitempty"` // Search query for name or license
	IncludeDeleted bool    `json:"includeDeleted"`
}

// DriverListResponse represents the paginated response for driver listing
type DriverListResponse struct {
	Items    []DriverResponse `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int64            `json:"total"`
}

// DriverResponse represents the response for a single driver
type DriverResponse struct {
	ID           uuid.UUID  `json:"id"`
	FullName     string     `json:"fullName"`
	Phone        *string    `json:"phone,omitempty"`
	LicenseNo    string     `json:"licenseNo"`
	LicenseClass string     `json:"licenseClass"`
	Photo        *string    `json:"photo,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
}

// ToResponse converts a Driver model to DriverResponse
func (d *Driver) ToResponse() DriverResponse {
	return DriverResponse{
		ID:           d.ID,
		FullName:     d.FullName,
		Phone:        d.Phone,
		LicenseNo:    d.LicenseNo,
		LicenseClass: d.LicenseClass,
		Photo:        d.Photo,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		DeletedAt:    d.DeletedAt,
	}
}

// FromDriverCreateRequest creates a Driver from CreateDriverRequest
func FromDriverCreateRequest(req CreateDriverRequest) *Driver {
	now := time.Now()
	return &Driver{
		FullName:     req.FullName,
		Phone:        req.Phone,
		LicenseNo:    req.LicenseNo,
		LicenseClass: string(req.LicenseClass),
		Photo:        req.Photo,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UpdateFromRequest updates a Driver from UpdateDriverRequest
func (d *Driver) UpdateFromRequest(req UpdateDriverRequest) {
	if req.FullName != nil {
		d.FullName = *req.FullName
	}
	if req.Phone != nil {
		d.Phone = req.Phone
	}
	if req.LicenseNo != nil {
		d.LicenseNo = *req.LicenseNo
	}
	if req.LicenseClass != nil {
		d.LicenseClass = string(*req.LicenseClass)
	}
	if req.Photo != nil {
		d.Photo = req.Photo
	}
	d.UpdatedAt = time.Now()
}
