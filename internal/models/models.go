package models

import "time"

// Client represents a company client
type Client struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Phone   string         `json:"phone"`
	Email   *string        `json:"email,omitempty"`
	Note    *string        `json:"note,omitempty"`
	Objects []ClientObject `json:"objects"`
}

// ClientObject represents a physical location/address for a client
type ClientObject struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	EquipmentID *string  `json:"equipment_id,omitempty"`
	Photos      []string `json:"photos,omitempty"`
}

// Equipment represents waste bins and containers
type Equipment struct {
	ID           string   `json:"id"`
	Number       *string  `json:"number,omitempty"`
	Type         string   `json:"type"`
	Volume       int      `json:"volume"`
	Condition    string   `json:"condition"`
	LocationType string   `json:"location_type"`
	Location     string   `json:"location"`
	Photos       []string `json:"photos"`
}

// Transport represents vehicles
type Transport struct {
	ID           string   `json:"id"`
	Brand        string   `json:"brand"`
	Model        string   `json:"model"`
	LicensePlate string   `json:"license_plate"`
	Status       string   `json:"status"`
	Capacity     int      `json:"capacity"`
	DriverID     *string  `json:"driver_id,omitempty"`
	EquipmentID  *string  `json:"equipment_id,omitempty"`
	Photos       []string `json:"photos"`
}

// Driver represents personnel
type Driver struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Phone         string   `json:"phone"`
	LicenseNumber string   `json:"license_number"`
	StartDate     string   `json:"start_date"`
	Photos        []string `json:"photos,omitempty"`
}

// Order represents waste collection requests
type Order struct {
	ID          string    `json:"id"`
	ObjectID    string    `json:"object_id"`
	Date        string    `json:"date"`
	Priority    string    `json:"priority"`
	TransportID *string   `json:"transport_id,omitempty"`
	Note        *string   `json:"note,omitempty"`
	Status      string    `json:"status"`
	CompletedAt *string   `json:"completed_at,omitempty"`
	Photos      []string  `json:"photos,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
