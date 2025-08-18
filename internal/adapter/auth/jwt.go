package auth

import (
	"fmt"
	"time"

	"eco-van-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Default TTLs
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour // 30 days
)

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID       `json:"sub"`
	Role   models.UserRole `json:"role"`
	Type   string          `json:"typ"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:  []byte(secretKey),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// NewDefaultJWTManager creates a JWT manager with default TTLs
func NewDefaultJWTManager(secretKey string) *JWTManager {
	return NewJWTManager(secretKey, AccessTokenTTL, RefreshTokenTTL)
}

// GenerateAccessToken generates a new access token
func (j *JWTManager) GenerateAccessToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTManager) GenerateRefreshToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token specifically
func (j *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeAccess {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token specifically
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	return claims, nil
}

// GetTokenExpiration returns the expiration time for a token
func (j *JWTManager) GetTokenExpiration(tokenType string) time.Duration {
	switch tokenType {
	case TokenTypeAccess:
		return j.accessTTL
	case TokenTypeRefresh:
		return j.refreshTTL
	default:
		return 0
	}
}

// GetAccessTokenTTL returns the access token TTL in seconds
func (j *JWTManager) GetAccessTokenTTL() int {
	return int(j.accessTTL.Seconds())
}
