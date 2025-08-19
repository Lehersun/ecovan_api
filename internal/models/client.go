package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateClientRequest represents the request to create a new client
type CreateClientRequest struct {
	Name  string  `json:"name" validate:"required,min=1,max=255"`
	TaxID *string `json:"tax_id,omitempty" validate:"omitempty,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	Notes *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateClientRequest represents the request to update an existing client
type UpdateClientRequest struct {
	Name  string  `json:"name" validate:"required,min=1,max=255"`
	TaxID *string `json:"tax_id,omitempty" validate:"omitempty,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	Notes *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// ClientListRequest represents the request to list clients with filtering and pagination
type ClientListRequest struct {
	Page           int    `json:"page" validate:"min=1"`
	PageSize       int    `json:"pageSize" validate:"min=1,max=100"`
	Query          string `json:"q" validate:"max=255"`
	IncludeDeleted bool   `json:"includeDeleted"`
}

// ClientListResponse represents the paginated response for listing clients
type ClientListResponse struct {
	Items    []Client `json:"items"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	Total    int64    `json:"total"`
}

// ClientResponse represents a single client response
type ClientResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	TaxID     *string    `json:"tax_id,omitempty"`
	Email     *string    `json:"email,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToResponse converts a Client model to ClientResponse
func (c *Client) ToResponse() ClientResponse {
	return ClientResponse{
		ID:        c.ID,
		Name:      c.Name,
		TaxID:     c.TaxID,
		Email:     c.Email,
		Phone:     c.Phone,
		Notes:     c.Notes,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: c.DeletedAt,
	}
}

// FromCreateRequest creates a new Client from CreateClientRequest
func FromCreateRequest(req CreateClientRequest) Client {
	return Client{
		Name:  req.Name,
		TaxID: req.TaxID,
		Email: req.Email,
		Phone: req.Phone,
		Notes: req.Notes,
	}
}

// UpdateFromRequest updates a Client from UpdateClientRequest
func (c *Client) UpdateFromRequest(req UpdateClientRequest) {
	c.Name = req.Name
	c.TaxID = req.TaxID
	c.Email = req.Email
	c.Phone = req.Phone
	c.Notes = req.Notes
}
