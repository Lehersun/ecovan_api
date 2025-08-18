package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the user's role in the system
type UserRole string

const (
	UserRoleAdmin     UserRole = "ADMIN"
	UserRoleDispatcher UserRole = "DISPATCHER"
	UserRoleDriver    UserRole = "DRIVER"
	UserRoleViewer    UserRole = "VIEWER"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the response from authentication endpoints
type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ValidateUserRole checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleDispatcher, UserRoleDriver, UserRoleViewer:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// ValidateCreateUserRequest validates the create user request
func (req *CreateUserRequest) Validate() error {
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	if !req.Role.IsValid() {
		return errors.New("invalid role")
	}
	
	return nil
}

// ValidateLoginRequest validates the login request
func (req *LoginRequest) Validate() error {
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	
	return nil
}

// ValidateRefreshRequest validates the refresh request
func (req *RefreshRequest) Validate() error {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return errors.New("refresh token is required")
	}
	
	return nil
}

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
