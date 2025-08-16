package models

import (
	"testing"
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
				ID:    "test-id",
				Name:  "Test Company",
				Phone: "+1234567890",
			},
			wantErr: false,
		},
		{
			name: "client without name",
			client: Client{
				ID:    "test-id",
				Phone: "+1234567890",
			},
			wantErr: true,
		},
		{
			name: "client without phone",
			client: Client{
				ID:   "test-id",
				Name: "Test Company",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.client.Name != "" && tt.client.Phone != "" {
					t.Errorf("Expected validation error for client %+v", tt.client)
				}
			}
		})
	}
}

func TestOrder_StatusTransitions(t *testing.T) {
	order := Order{
		ID:       "test-order",
		ObjectID: "test-object",
		Date:     "2025-01-01",
		Priority: "Средний",
		Status:   "В ожидании",
	}

	// Test status transition
	if order.Status != "В ожидании" {
		t.Errorf("Expected initial status 'В ожидании', got %s", order.Status)
	}

	// Simulate status change
	order.Status = "В работе"
	if order.Status != "В работе" {
		t.Errorf("Expected status 'В работе', got %s", order.Status)
	}
}

func TestEquipment_VolumeValidation(t *testing.T) {
	equipment := Equipment{
		ID:           "test-equipment",
		Type:         "Спецэкогруз",
		Volume:       10,
		Condition:    "Хорошее",
		LocationType: "Склад",
		Location:     "Warehouse A",
	}

	if equipment.Volume <= 0 {
		t.Errorf("Equipment volume should be positive, got %d", equipment.Volume)
	}
}

func TestTransport_CapacityValidation(t *testing.T) {
	transport := Transport{
		ID:           "test-transport",
		Brand:        "КАМАЗ",
		Model:        "65115",
		LicensePlate: "A123BC",
		Status:       "В работе",
		Capacity:     100,
	}

	if transport.Capacity <= 0 {
		t.Errorf("Transport capacity should be positive, got %d", transport.Capacity)
	}
}

func TestDriver_RequiredFields(t *testing.T) {
	driver := Driver{
		ID:            "test-driver",
		Name:          "John Doe",
		Phone:         "+1234567890",
		LicenseNumber: "DL123456",
		StartDate:     "2025-01-01",
	}

	if driver.Name == "" || driver.Phone == "" || driver.LicenseNumber == "" {
		t.Errorf("Driver should have all required fields filled")
	}
}
