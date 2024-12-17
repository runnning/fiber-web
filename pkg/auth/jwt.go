package auth

import (
	"errors"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenType represents the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"`
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken         string        `json:"access_token"`
	RefreshToken        string        `json:"refresh_token"`
	AccessTokenExpires  time.Duration `json:"access_token_expires"`
	RefreshTokenExpires time.Duration `json:"refresh_token_expires"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secretKey:          cfg.JWT.SecretKey,
		accessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		refreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (m *JWTManager) GenerateTokenPair(userID uint64, username, role string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := m.generateToken(userID, username, role, AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := m.generateToken(userID, username, role, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		AccessTokenExpires:  m.accessTokenExpiry,
		RefreshTokenExpires: m.refreshTokenExpiry,
	}, nil
}

// generateToken generates a new JWT token
func (m *JWTManager) generateToken(userID uint64, username, role string, tokenType TokenType) (string, error) {
	var expiry time.Duration
	if tokenType == AccessToken {
		expiry = m.accessTokenExpiry
	} else {
		expiry = m.refreshTokenExpiry
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     string(tokenType),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns its claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		logger.Error("Failed to parse token", zap.Error(err))
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken validates a refresh token and generates a new access token
func (m *JWTManager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != string(RefreshToken) {
		return nil, errors.New("invalid token type")
	}

	return m.GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
}
