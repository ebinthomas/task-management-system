package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager handles JWT token operations
type TokenManager struct {
	secretKey     []byte
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// NewTokenManager creates a new token manager
func NewTokenManager(secretKey []byte, issuer string) *TokenManager {
	return &TokenManager{
		secretKey:     secretKey,
		issuer:        issuer,
		accessExpiry:  15 * time.Minute,  // Access tokens expire in 15 minutes
		refreshExpiry: 7 * 24 * time.Hour, // Refresh tokens expire in 7 days
	}
}

// CreateTokenPair generates a new access and refresh token pair
func (tm *TokenManager) CreateTokenPair(userID string, roles []string) (*TokenPair, error) {
	// Create access token
	accessToken, err := tm.createToken(userID, roles, tm.accessExpiry)
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshToken, err := tm.createRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tm.accessExpiry.Seconds()),
	}, nil
}

// createToken generates a new JWT token
func (tm *TokenManager) createToken(userID string, roles []string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.issuer,
			Subject:   userID,
		},
		UserID: userID,
		Roles:  roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// createRefreshToken generates a new refresh token
func (tm *TokenManager) createRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(tm.refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    tm.issuer,
		Subject:   userID,
		ID:        generateTokenID(), // Unique ID for token revocation
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// RefreshTokens validates a refresh token and issues new token pair
func (tm *TokenManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// Parse the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// Check if token is revoked (implement token blacklist check here)
		if isTokenRevoked(claims.ID) {
			return nil, ErrInvalidToken
		}

		// Get user roles from your user service/database
		roles, err := getUserRoles(claims.Subject)
		if err != nil {
			return nil, err
		}

		// Create new token pair
		return tm.CreateTokenPair(claims.Subject, roles)
	}

	return nil, ErrInvalidToken
}

// ValidateToken validates a JWT token and returns its claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tm.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != tm.issuer {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}

// Helper functions (to be implemented based on your storage solution)
func generateTokenID() string {
	// Implement unique ID generation (e.g., UUID)
	return "unique-token-id"
}

func isTokenRevoked(tokenID string) bool {
	// Implement token revocation check (e.g., Redis blacklist)
	return false
}

func getUserRoles(userID string) ([]string, error) {
	// Implement role lookup from your database
	return []string{"user"}, nil
} 