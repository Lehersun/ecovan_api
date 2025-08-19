package models

import (
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
	ClientID            uuid.UUID  `json:"client_id" db:"client_id"`
	ObjectID            uuid.UUID  `json:"object_id" db:"object_id"`
	ScheduledDate       string     `json:"scheduled_date" db:"scheduled_date"`
	ScheduledWindowFrom *string    `json:"scheduled_window_from,omitempty" db:"scheduled_window_from"`
	ScheduledWindowTo   *string    `json:"scheduled_window_to,omitempty" db:"scheduled_window_to"`
	Status              string     `json:"status" db:"status"`
	TransportID         *uuid.UUID `json:"transport_id,omitempty" db:"transport_id"`
	Notes               *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
