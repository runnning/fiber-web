package auth

import (
	"errors"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("无效的令牌")
	ErrExpiredToken = errors.New("令牌已过期")
)

var (
	jwtManager     *JWTManager
	jwtManagerOnce sync.Once
)

// GetJWTManager 获取JWTManager的单例实例
func GetJWTManager() *JWTManager {
	return jwtManager
}

// InitJWTManager 初始化JWTManager单例
func InitJWTManager(cfg *config.JWTConfig) {
	jwtManagerOnce.Do(func() {
		jwtManager = NewJWTManager(cfg)
	})
}

// TokenType 表示令牌类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims 表示JWT的声明内容
type Claims struct {
	jwt.RegisteredClaims
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"`
}

// TokenPair 表示访问令牌和刷新令牌的配对
type TokenPair struct {
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	AccessTokenExpires  int64  `json:"access_token_expires"`  // 过期时间（秒）
	RefreshTokenExpires int64  `json:"refresh_token_expires"` // 过期时间（秒）
}

// JWTManager 处理JWT令牌的操作
type JWTManager struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTManager 创建一个新的JWT管理器
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:          cfg.SecretKey,
		accessTokenExpiry:  cfg.AccessTokenExpiry,
		refreshTokenExpiry: cfg.RefreshTokenExpiry,
	}
}

// GenerateTokenPair 生成新的访问令牌和刷新令牌配对
func (m *JWTManager) GenerateTokenPair(userID uint64, username, role string) (*TokenPair, error) {
	// 生成访问令牌
	accessToken, err := m.generateToken(userID, username, role, AccessToken)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := m.generateToken(userID, username, role, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		AccessTokenExpires:  int64(m.accessTokenExpiry.Seconds()),
		RefreshTokenExpires: int64(m.refreshTokenExpiry.Seconds()),
	}, nil
}

// generateToken 生成一个新的JWT令牌
func (m *JWTManager) generateToken(userID uint64, username, role string, tokenType TokenType) (string, error) {
	expiry := m.accessTokenExpiry
	if tokenType == RefreshToken {
		expiry = m.refreshTokenExpiry
	}

	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     string(tokenType),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken 验证JWT令牌并返回其声明内容
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// 验证令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		logger.Error("解析令牌失败", zap.Error(err))
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

// RefreshToken 验证刷新令牌并生成新的令牌配对
func (m *JWTManager) RefreshToken(refreshToken string) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != string(RefreshToken) {
		return nil, ErrInvalidToken
	}

	// 生成新的令牌对
	return m.GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
}
