package auth

import (
	"testing"
	"time"

	"eco-van-api/internal/models"

	"github.com/google/uuid"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	accessTTL := 15 * time.Minute
	refreshTTL := 30 * 24 * time.Hour

	manager := NewJWTManager(secretKey, accessTTL, refreshTTL)

	if manager == nil {
		t.Fatal("Expected JWT manager to be created")
	}

	if manager.accessTTL != accessTTL {
		t.Errorf("Expected access TTL %v, got %v", accessTTL, manager.accessTTL)
	}

	if manager.refreshTTL != refreshTTL {
		t.Errorf("Expected refresh TTL %v, got %v", refreshTTL, manager.refreshTTL)
	}
}

func TestNewDefaultJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	if manager == nil {
		t.Fatal("Expected JWT manager to be created")
	}

	if manager.accessTTL != AccessTokenTTL {
		t.Errorf("Expected access TTL %v, got %v", AccessTokenTTL, manager.accessTTL)
	}

	if manager.refreshTTL != RefreshTokenTTL {
		t.Errorf("Expected refresh TTL %v, got %v", RefreshTokenTTL, manager.refreshTTL)
	}
}

func TestGenerateAccessToken(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	token, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty access token")
	}

	// Validate the token
	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	if claims.Type != TokenTypeAccess {
		t.Errorf("Expected token type %s, got %s", TokenTypeAccess, claims.Type)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	token, err := manager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Validate the token
	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	if claims.Type != TokenTypeRefresh {
		t.Errorf("Expected token type %s, got %s", TokenTypeRefresh, claims.Type)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	// Test with invalid token
	_, err := manager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error when validating invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Create manager with one secret
	manager1 := NewDefaultJWTManager("secret1")

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	// Generate token with first manager
	token, err := manager1.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	manager2 := NewDefaultJWTManager("secret2")
	_, err = manager2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret")
	}
}

func TestValidateAccessToken_RefreshToken(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	// Generate refresh token
	token, err := manager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Try to validate as access token
	_, err = manager.ValidateAccessToken(token)
	if err == nil {
		t.Error("Expected error when validating refresh token as access token")
	}
}

func TestValidateRefreshToken_AccessToken(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	// Generate access token
	token, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Try to validate as refresh token
	_, err = manager.ValidateRefreshToken(token)
	if err == nil {
		t.Error("Expected error when validating access token as refresh token")
	}
}

func TestGetTokenExpiration(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	accessTTL := manager.GetTokenExpiration(TokenTypeAccess)
	if accessTTL != AccessTokenTTL {
		t.Errorf("Expected access TTL %v, got %v", AccessTokenTTL, accessTTL)
	}

	refreshTTL := manager.GetTokenExpiration(TokenTypeRefresh)
	if refreshTTL != RefreshTokenTTL {
		t.Errorf("Expected refresh TTL %v, got %v", RefreshTokenTTL, refreshTTL)
	}

	unknownTTL := manager.GetTokenExpiration("unknown")
	if unknownTTL != 0 {
		t.Errorf("Expected unknown TTL 0, got %v", unknownTTL)
	}
}

func TestGetAccessTokenTTL(t *testing.T) {
	secretKey := "test-secret-key"
	manager := NewDefaultJWTManager(secretKey)

	expectedSeconds := int(AccessTokenTTL.Seconds())
	actualSeconds := manager.GetAccessTokenTTL()

	if actualSeconds != expectedSeconds {
		t.Errorf("Expected %d seconds, got %d", expectedSeconds, actualSeconds)
	}
}
