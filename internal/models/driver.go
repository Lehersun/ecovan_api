package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverLicenseClass represents the class of driver's license
type DriverLicenseClass string

const (
	DriverLicenseClassA   DriverLicenseClass = "A"   // Motorcycle
	DriverLicenseClassA1  DriverLicenseClass = "A1"  // Light motorcycle
	DriverLicenseClassB   DriverLicenseClass = "B"   // Car
	DriverLicenseClassB1  DriverLicenseClass = "B1"  // Light car
	DriverLicenseClassC   DriverLicenseClass = "C"   // Truck
	DriverLicenseClassC1  DriverLicenseClass = "C1"  // Light truck
	DriverLicenseClassD   DriverLicenseClass = "D"   // Bus
	DriverLicenseClassD1  DriverLicenseClass = "D1"  // Light bus
	DriverLicenseClassBE  DriverLicenseClass = "BE"  // Car with trailer
	DriverLicenseClassB1E DriverLicenseClass = "B1E" // Light car with trailer
	DriverLicenseClassCE  DriverLicenseClass = "CE"  // Truck with trailer
	DriverLicenseClassC1E DriverLicenseClass = "C1E" // Light truck with trailer
	DriverLicenseClassDE  DriverLicenseClass = "DE"  // Bus with trailer
	DriverLicenseClassD1E DriverLicenseClass = "D1E" // Light bus with trailer
)

// CreateDriverRequest represents the request to create a new driver
type CreateDriverRequest struct {
	FullName       string               `json:"fullName" validate:"required,min=2,max=100"`
	Phone          *string              `json:"phone,omitempty" validate:"omitempty,max=20"`
	LicenseNo      *string              `json:"licenseNo,omitempty" validate:"omitempty,min=5,max=20"`
	LicenseClasses []DriverLicenseClass `json:"licenseClasses,omitempty" validate:"omitempty,dive,oneof=A A1 B B1 C C1 D D1 BE B1E CE C1E DE D1E"` //nolint:lll // long license validation enum
	Photo          *string              `json:"photo,omitempty" validate:"omitempty,max=500"`
}

// UpdateDriverRequest represents the request to update an existing driver
type UpdateDriverRequest struct {
	FullName       *string              `json:"fullName,omitempty" validate:"omitempty,min=2,max=100"`
	Phone          *string              `json:"phone,omitempty" validate:"omitempty,max=20"`
	LicenseNo      *string              `json:"licenseNo,omitempty" validate:"omitempty,min=5,max=20"`
	LicenseClasses []DriverLicenseClass `json:"licenseClasses,omitempty" validate:"omitempty,dive,oneof=A A1 B B1 C C1 D D1 BE B1E CE C1E DE D1E"` //nolint:lll // long license validation enum
	Photo          *string              `json:"photo,omitempty" validate:"omitempty,max=500"`
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
	ID             uuid.UUID  `json:"id"`
	FullName       string     `json:"fullName"`
	Phone          *string    `json:"phone,omitempty"`
	LicenseNo      *string    `json:"licenseNo,omitempty"`
	LicenseClasses []string   `json:"licenseClasses,omitempty"`
	Photo          *string    `json:"photo,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

// ToResponse converts a Driver model to DriverResponse
func (d *Driver) ToResponse() DriverResponse {
	return DriverResponse{
		ID:             d.ID,
		FullName:       d.FullName,
		Phone:          d.Phone,
		LicenseNo:      d.LicenseNo,
		LicenseClasses: d.LicenseClasses,
		Photo:          d.Photo,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
	}
}

// FromDriverCreateRequest creates a Driver from CreateDriverRequest
func FromDriverCreateRequest(req CreateDriverRequest) *Driver {
	now := time.Now()
	driver := &Driver{
		FullName:  req.FullName,
		Phone:     req.Phone,
		Photo:     req.Photo,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set license fields if provided
	if req.LicenseNo != nil {
		driver.LicenseNo = req.LicenseNo
	}
	if len(req.LicenseClasses) > 0 {
		driver.LicenseClasses = make([]string, len(req.LicenseClasses))
		for i, class := range req.LicenseClasses {
			driver.LicenseClasses[i] = string(class)
		}
	}

	return driver
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
		d.LicenseNo = req.LicenseNo
	}
	if len(req.LicenseClasses) > 0 {
		d.LicenseClasses = make([]string, len(req.LicenseClasses))
		for i, class := range req.LicenseClasses {
			d.LicenseClasses[i] = string(class)
		}
	}
	if req.Photo != nil {
		d.Photo = req.Photo
	}
	d.UpdatedAt = time.Now()
}
