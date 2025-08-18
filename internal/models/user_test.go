package models

import (
	"testing"
)

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectValid bool
	}{
		{
			name:        "Valid email with domain",
			email:       "user@example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with subdomain",
			email:       "user@sub.example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with numbers",
			email:       "user123@example123.com",
			expectValid: true,
		},
		{
			name:        "Valid email with dots in local part",
			email:       "user.name@example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with plus sign",
			email:       "user+tag@example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with underscore",
			email:       "user_name@example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with dash",
			email:       "user-name@example.com",
			expectValid: true,
		},
		{
			name:        "Valid email with percent",
			email:       "user%name@example.com",
			expectValid: true,
		},
		{
			name:        "Invalid email - missing @",
			email:       "userexample.com",
			expectValid: false,
		},
		{
			name:        "Invalid email - missing domain",
			email:       "user@",
			expectValid: false,
		},
		{
			name:        "Invalid email - missing local part",
			email:       "@example.com",
			expectValid: false,
		},
		{
			name:        "Invalid email - invalid characters",
			email:       "user@example..com",
			expectValid: false,
		},
		{
			name:        "Invalid email - spaces",
			email:       "user @example.com",
			expectValid: false,
		},
		{
			name:        "Invalid email - empty string",
			email:       "",
			expectValid: false,
		},
		{
			name:        "Invalid email - just @",
			email:       "@",
			expectValid: false,
		},
		{
			name:        "Invalid email - invalid TLD",
			email:       "user@example.c",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateUserRequest{
				Email:    tt.email,
				Password: "validpassword123",
				Role:     UserRoleViewer,
			}

			err := req.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("Expected valid email '%s' but got error: %v", tt.email, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid email '%s' but got no error", tt.email)
			}
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectValid bool
	}{
		{
			name:        "Valid password - 8 characters",
			password:    "pass1234",
			expectValid: true,
		},
		{
			name:        "Valid password - longer than 8",
			password:    "verylongpassword123",
			expectValid: true,
		},
		{
			name:        "Invalid password - too short",
			password:    "pass",
			expectValid: false,
		},
		{
			name:        "Invalid password - 7 characters",
			password:    "pass123",
			expectValid: false,
		},
		{
			name:        "Invalid password - empty string",
			password:    "",
			expectValid: false,
		},
		{
			name:        "Invalid password - only spaces",
			password:    "   ",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateUserRequest{
				Email:    "test@example.com",
				Password: tt.password,
				Role:     UserRoleViewer,
			}

			err := req.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("Expected valid password '%s' but got error: %v", tt.password, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid password '%s' but got no error", tt.password)
			}
		})
	}
}

func TestRoleValidation(t *testing.T) {
	tests := []struct {
		name        string
		role        UserRole
		expectValid bool
	}{
		{
			name:        "Valid role - ADMIN",
			role:        UserRoleAdmin,
			expectValid: true,
		},
		{
			name:        "Valid role - DISPATCHER",
			role:        UserRoleDispatcher,
			expectValid: true,
		},
		{
			name:        "Valid role - DRIVER",
			role:        UserRoleDriver,
			expectValid: true,
		},
		{
			name:        "Valid role - VIEWER",
			role:        UserRoleViewer,
			expectValid: true,
		},
		{
			name:        "Invalid role - empty string",
			role:        "",
			expectValid: false,
		},
		{
			name:        "Invalid role - random string",
			role:        "RANDOM",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateUserRequest{
				Email:    "test@example.com",
				Password: "validpassword123",
				Role:     tt.role,
			}

			err := req.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("Expected valid role '%s' but got error: %v", tt.role, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid role '%s' but got no error", tt.role)
			}
		})
	}
}

func TestCreateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateUserRequest
		expectValid bool
	}{
		{
			name: "Valid request",
			req: CreateUserRequest{
				Email:    "user@example.com",
				Password: "password123",
				Role:     UserRoleViewer,
			},
			expectValid: true,
		},
		{
			name: "Invalid - missing email",
			req: CreateUserRequest{
				Email:    "",
				Password: "password123",
				Role:     UserRoleViewer,
			},
			expectValid: false,
		},
		{
			name: "Invalid - missing password",
			req: CreateUserRequest{
				Email:    "user@example.com",
				Password: "",
				Role:     UserRoleViewer,
			},
			expectValid: false,
		},
		{
			name: "Invalid - missing role",
			req: CreateUserRequest{
				Email:    "user@example.com",
				Password: "password123",
				Role:     "",
			},
			expectValid: false,
		},
		{
			name: "Invalid - all fields missing",
			req: CreateUserRequest{
				Email:    "",
				Password: "",
				Role:     "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("Expected valid request but got error: %v", err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid request but got no error")
			}
		})
	}
}
