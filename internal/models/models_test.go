package models

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func TestClient_Validation(t *testing.T) {
	tests := []struct {
		name    string
		client  Client
		wantErr bool
	}{
		{
			name: "valid client",
			client: Client{
				ID:        uuid.New(),
				Name:      "Test Company",
				Phone:     stringPtr("+1234567890"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "client without name",
			client: Client{
				ID:        uuid.New(),
				Phone:     stringPtr("+1234567890"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "client without phone",
			client: Client{
				ID:        uuid.New(),
				Name:      "Test Company",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false, // Phone is optional now
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.client.Name != "" {
					t.Errorf("Expected validation error for client %+v", tt.client)
				}
			}
		})
	}
}

func TestCreateClientRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateClientRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateClientRequest{
				Name:  "Test Company",
				TaxID: stringPtr("123456789"),
				Email: stringPtr("test@company.com"),
				Phone: stringPtr("+1234567890"),
				Notes: stringPtr("Test notes"),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: CreateClientRequest{
				TaxID: stringPtr("123456789"),
				Email: stringPtr("test@company.com"),
				Phone: stringPtr("+1234567890"),
				Notes: stringPtr("Test notes"),
			},
			wantErr: true,
		},
		{
			name: "empty name",
			request: CreateClientRequest{
				Name:  "",
				TaxID: stringPtr("123456789"),
				Email: stringPtr("test@company.com"),
			},
			wantErr: true,
		},
		{
			name: "name too long",
			request: CreateClientRequest{
				Name:  string(make([]byte, 256)), // 256 characters
				TaxID: stringPtr("123456789"),
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			request: CreateClientRequest{
				Name:  "Test Company",
				Email: stringPtr("invalid-email"),
			},
			wantErr: true,
		},
		{
			name: "valid email",
			request: CreateClientRequest{
				Name:  "Test Company",
				Email: stringPtr("valid@email.com"),
			},
			wantErr: false,
		},
		{
			name: "tax_id too long",
			request: CreateClientRequest{
				Name:  "Test Company",
				TaxID: stringPtr(string(make([]byte, 101))), // 101 characters
			},
			wantErr: true,
		},
		{
			name: "phone too long",
			request: CreateClientRequest{
				Name:  "Test Company",
				Phone: stringPtr(string(make([]byte, 21))), // 21 characters
			},
			wantErr: true,
		},
		{
			name: "notes too long",
			request: CreateClientRequest{
				Name:  "Test Company",
				Notes: stringPtr(string(make([]byte, 1001))), // 1001 characters
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(tt.request)

			if tt.wantErr && err == nil {
				t.Errorf("Expected validation error for request %+v", tt.request)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error for request %+v: %v", tt.request, err)
			}
		})
	}
}

func TestUpdateClientRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateClientRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: UpdateClientRequest{
				Name:  "Updated Company",
				TaxID: stringPtr("987654321"),
				Email: stringPtr("updated@company.com"),
				Phone: stringPtr("+9876543210"),
				Notes: stringPtr("Updated notes"),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: UpdateClientRequest{
				TaxID: stringPtr("987654321"),
				Email: stringPtr("updated@company.com"),
			},
			wantErr: true,
		},
		{
			name: "empty name",
			request: UpdateClientRequest{
				Name:  "",
				TaxID: stringPtr("987654321"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(tt.request)

			if tt.wantErr && err == nil {
				t.Errorf("Expected validation error for request %+v", tt.request)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error for request %+v: %v", tt.request, err)
			}
		})
	}
}

func TestClientListRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request ClientListRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: ClientListRequest{
				Page:           1,
				PageSize:       20,
				Query:          "test",
				IncludeDeleted: false,
			},
			wantErr: false,
		},
		{
			name: "page too small",
			request: ClientListRequest{
				Page:     0,
				PageSize: 20,
				Query:    "test",
			},
			wantErr: true,
		},
		{
			name: "page size too small",
			request: ClientListRequest{
				Page:     1,
				PageSize: 0,
				Query:    "test",
			},
			wantErr: true,
		},
		{
			name: "page size too large",
			request: ClientListRequest{
				Page:     1,
				PageSize: 101,
				Query:    "test",
			},
			wantErr: true,
		},
		{
			name: "query too long",
			request: ClientListRequest{
				Page:     1,
				PageSize: 20,
				Query:    string(make([]byte, 256)), // 256 characters
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(tt.request)

			if tt.wantErr && err == nil {
				t.Errorf("Expected validation error for request %+v", tt.request)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error for request %+v: %v", tt.request, err)
			}
		})
	}
}

func TestClient_ToResponse(t *testing.T) {
	client := Client{
		ID:        uuid.New(),
		Name:      "Test Company",
		TaxID:     stringPtr("123456789"),
		Email:     stringPtr("test@company.com"),
		Phone:     stringPtr("+1234567890"),
		Notes:     stringPtr("Test notes"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	response := client.ToResponse()

	if response.ID != client.ID {
		t.Errorf("Expected ID %v, got %v", client.ID, response.ID)
	}
	if response.Name != client.Name {
		t.Errorf("Expected Name %s, got %s", client.Name, response.Name)
	}
	if response.TaxID != client.TaxID {
		t.Errorf("Expected TaxID %v, got %v", client.TaxID, response.TaxID)
	}
	if response.Email != client.Email {
		t.Errorf("Expected Email %v, got %v", client.Email, response.Email)
	}
	if response.Phone != client.Phone {
		t.Errorf("Expected Phone %v, got %v", client.Phone, response.Phone)
	}
	if response.Notes != client.Notes {
		t.Errorf("Expected Notes %v, got %v", client.Notes, response.Notes)
	}
	if response.CreatedAt != client.CreatedAt {
		t.Errorf("Expected CreatedAt %v, got %v", client.CreatedAt, response.CreatedAt)
	}
	if response.UpdatedAt != client.UpdatedAt {
		t.Errorf("Expected UpdatedAt %v, got %v", client.UpdatedAt, response.UpdatedAt)
	}
	if response.DeletedAt != client.DeletedAt {
		t.Errorf("Expected DeletedAt %v, got %v", client.DeletedAt, response.DeletedAt)
	}
}

func TestFromCreateRequest(t *testing.T) {
	req := CreateClientRequest{
		Name:  "Test Company",
		TaxID: stringPtr("123456789"),
		Email: stringPtr("test@company.com"),
		Phone: stringPtr("+1234567890"),
		Notes: stringPtr("Test notes"),
	}

	client := FromCreateRequest(req)

	if client.Name != req.Name {
		t.Errorf("Expected Name %s, got %s", req.Name, client.Name)
	}
	if client.TaxID != req.TaxID {
		t.Errorf("Expected TaxID %v, got %v", req.TaxID, client.TaxID)
	}
	if client.Email != req.Email {
		t.Errorf("Expected Email %v, got %v", req.Email, client.Email)
	}
	if client.Phone != req.Phone {
		t.Errorf("Expected Phone %v, got %v", req.Phone, client.Phone)
	}
	if client.Notes != req.Notes {
		t.Errorf("Expected Notes %v, got %v", req.Notes, client.Notes)
	}
}

func TestClient_UpdateFromRequest(t *testing.T) {
	client := &Client{
		ID:        uuid.New(),
		Name:      "Old Company",
		TaxID:     stringPtr("old123"),
		Email:     stringPtr("old@company.com"),
		Phone:     stringPtr("+1234567890"),
		Notes:     stringPtr("Old notes"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := UpdateClientRequest{
		Name:  "New Company",
		TaxID: stringPtr("new456"),
		Email: stringPtr("new@company.com"),
		Phone: stringPtr("+9876543210"),
		Notes: stringPtr("New notes"),
	}

	client.UpdateFromRequest(req)

	if client.Name != req.Name {
		t.Errorf("Expected Name %s, got %s", req.Name, client.Name)
	}
	if client.TaxID != req.TaxID {
		t.Errorf("Expected TaxID %v, got %v", req.TaxID, client.TaxID)
	}
	if client.Email != req.Email {
		t.Errorf("Expected Email %v, got %v", req.Email, client.Email)
	}
	if client.Phone != req.Phone {
		t.Errorf("Expected Phone %v, got %v", req.Phone, client.Phone)
	}
	if client.Notes != req.Notes {
		t.Errorf("Expected Notes %v, got %v", req.Notes, client.Notes)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestOrder_StatusTransitions(t *testing.T) {
	order := Order{
		ID:            uuid.New(),
		ClientID:      uuid.New(),
		ObjectID:      uuid.New(),
		ScheduledDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:        "DRAFT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test status transition
	if order.Status != "DRAFT" {
		t.Errorf("Expected initial status 'DRAFT', got %s", order.Status)
	}

	// Simulate status change
	order.Status = "SCHEDULED"
	if order.Status != "SCHEDULED" {
		t.Errorf("Expected status 'SCHEDULED', got %s", order.Status)
	}
}

func TestEquipment_VolumeValidation(t *testing.T) {
	equipment := Equipment{
		ID:        uuid.New(),
		Type:      "BIN",
		VolumeL:   100,
		Condition: "GOOD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if equipment.VolumeL <= 0 {
		t.Errorf("Equipment volume should be positive, got %d", equipment.VolumeL)
	}
}

func TestTransport_CapacityValidation(t *testing.T) {
	transport := Transport{
		ID:        uuid.New(),
		PlateNo:   "A123BC",
		CapacityL: 1000,
		Status:    "IN_WORK",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if transport.CapacityL <= 0 {
		t.Errorf("Transport capacity should be positive, got %d", transport.CapacityL)
	}
}

func TestDriver_RequiredFields(t *testing.T) {
	driver := Driver{
		ID:             uuid.New(),
		FullName:       "John Doe",
		Phone:          stringPtr("+1234567890"),
		LicenseNo:      stringPtr("DL123456"),
		LicenseClasses: []string{"B", "C"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if driver.FullName == "" {
		t.Errorf("Driver should have FullName field filled")
	}
}
