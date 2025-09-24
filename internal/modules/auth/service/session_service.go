// internal/modules/auth/service/session_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/entity"
	authpb "tsu-self/internal/rpc/generated/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type SessionService struct {
	redis     *redis.Client
	jwtSecret []byte
	tokenTTL  time.Duration
	logger    log.Logger
}

type JWTClaims struct {
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"session_id"`
	jwt.RegisteredClaims
}

type SessionData struct {
	Token       string    `json:"token"`
	UserID      string    `json:"user_id"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	LastAccess  time.Time `json:"last_access"`
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent"`
}

func NewSessionService(redis *redis.Client, jwtSecret string, tokenTTL time.Duration, logger log.Logger) *SessionService {
	return &SessionService{
		redis:     redis,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
		logger:    logger,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, userInfo *entity.User, clientIP, userAgent string) (string, error) {
	// 1. 生成 SessionID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", err
	}

	// 2. 创建 JWT Claims
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Roles:     []string{}, // 从权限服务获取
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// 3. 生成 JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	// 4. 存储到 Redis
	sessionData := &SessionData{
		Token:       tokenString,
		UserID:      userInfo.ID,
		Permissions: []string{}, // TODO: 从 Keto 获取权限
		CreatedAt:   now,
		LastAccess:  now,
		ClientIP:    clientIP,
		UserAgent:   userAgent,
	}

	sessionJSON, _ := json.Marshal(sessionData)
	key := fmt.Sprintf("session:%s:%s", userInfo.ID, sessionID)

	if err := s.redis.Set(ctx, key, sessionJSON, s.tokenTTL).Err(); err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *SessionService) ValidateToken(ctx context.Context, tokenString string) (*authpb.ValidateTokenResponse, *xerrors.AppError) {
	// 1. 解析 JWT
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	// 2. 检查 Redis 中的 Session
	key := fmt.Sprintf("session:%s:%s", claims.UserID, claims.SessionID)
	sessionJSON, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &authpb.ValidateTokenResponse{Valid: false}, nil
		}
		return nil, xerrors.NewExternalServiceError("redis", err)
	}

	// 3. 更新最后访问时间
	var sessionData SessionData
	json.Unmarshal([]byte(sessionJSON), &sessionData)
	sessionData.LastAccess = time.Now()

	updatedJSON, _ := json.Marshal(sessionData)
	s.redis.Set(ctx, key, updatedJSON, s.tokenTTL)

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
	}, nil
}

func (s *SessionService) InvalidateSession(ctx context.Context, userID, sessionID string) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	return s.redis.Del(ctx, key).Err()
}

func (s *SessionService) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("session:%s:*", userID)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redis.Del(ctx, keys...).Err()
	}
	return nil
}

func (s *SessionService) generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
